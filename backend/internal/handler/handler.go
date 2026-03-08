package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/makimaki04/Calmly/internal/models"
	"github.com/makimaki04/Calmly/internal/repository"
	"github.com/makimaki04/Calmly/internal/service"
	"github.com/makimaki04/Calmly/pkg/contract"
	"go.uber.org/zap"
)

type Handler struct {
	service *service.Service
	flow    *service.FlowService
	logger  *zap.Logger
}

func NewHandler(svc *service.Service, flow *service.FlowService, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		flow:    flow,
		logger:  logger.With(zap.String("component", "http")),
	}
}

const (
	MB          = 1 << 20
	MaxBodyAuth = 1 * MB
	MaxBodyJSON = 1 * MB
)

var mockUserID = uuid.New()

func parseJSONBody[T any](w http.ResponseWriter, r *http.Request, maxBytesRead int64, logger *zap.Logger) (T, error) {
	var req T

	r.Body = http.MaxBytesReader(w, r.Body, maxBytesRead)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			logger.Error("Decode request body failed", zap.Error(err))
			return req, fmt.Errorf("empty request body: %w", err)
		}

		logger.Error("Decode request body failed", zap.Error(err))
		return req, fmt.Errorf("decode json body: %w", err)
	}

	var extra any
	err := dec.Decode(&extra)
	switch err {
	case io.EOF:
		return req, nil
	case nil:
		logger.Error("request body contains additional JSON object")
		return req, fmt.Errorf("request body must contain a single JSON object")
	default:
		logger.Error("found trailing invalid data after JSON object", zap.Error(err))
		return req, fmt.Errorf("request body must contain a single JSON object")
	}
}

var ErrEmptyDumpText = errors.New("raw text is required")

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		zap.String("operation", "start_session"),
		zap.String("user_id", mockUserID.String()),
	)

	log.Info("User session started")

	req, err := parseJSONBody[contract.StartSessionRequest](w, r, MaxBodyJSON, log)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body", log)
		return
	}

	if req.Dump.RawText == nil {
		err = ErrEmptyDumpText
		log.Error("Start session failed", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "raw text is required", log)
		return
	}

	ctx := r.Context()
	analysis, err := h.flow.StartSession(ctx, mockUserID, *req.Dump.RawText)
	if err != nil {
		log.Error("Start session failed", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "failed to start session", log)
		return
	}

	resp := contract.StartSessionResponse{
		Analysis: models.ConvertToAnalysisDTO(analysis),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, resp, log)
}

func (h *Handler) SubmitAnswers(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		zap.String("operation", "submit_answers"),
		zap.String("user_id", mockUserID.String()),
	)

	log.Info("Submit answers started")

	dumpID, err := uuid.Parse(chi.URLParam(r, "dump_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "wrong url params", h.logger)
		return
	}

	req, err := parseJSONBody[contract.SubmitAnswersRequest](w, r, MaxBodyJSON, log)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body", log)
		return
	}

	log = log.With(zap.String("dump_id", dumpID.String()))

	var answers []models.Answer
	for _, a := range req.Answers {
		answer := models.Answer{
			QuestionID: a.QuestionID,
			Text:       a.Text,
		}

		answers = append(answers, answer)
	}

	dumpAnswers := models.DumpAnswers{
		DumpID:  dumpID,
		Answers: answers,
	}

	ctx := r.Context()
	plan, planItems, err := h.flow.SubmitAnswers(ctx, mockUserID, dumpAnswers)
	if err != nil {
		log.Error("Submit answers failed", zap.Error(err))
		respondWithError(w, mapFlowErrorToStatus(err), errorMessage(err), log)
		return
	}

	resp := contract.SubmitAnswersResponse{
		Plan:      models.ConvertToPlanDTO(plan),
		PlanItems: models.ConvertToPlanItemsDTO(planItems),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, resp, log)
}

func (h *Handler) RegeneratePlan(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		zap.String("operation", "regenerate_plan"),
		zap.String("user_id", mockUserID.String()),
	)

	log.Info("Regenerate plan started")

	dumpID, err := uuid.Parse(chi.URLParam(r, "dump_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "wrong url params", h.logger)
		return
	}

	req, err := parseJSONBody[contract.RegeneratePlanRequest](w, r, MaxBodyJSON, log)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body", log)
		return
	}

	log = log.With(zap.String("dump_id", dumpID.String()))

	fb := models.UserFeedback{
		DumpID: dumpID,
		Text:   req.Feedback,
	}

	ctx := r.Context()
	plan, items, err := h.flow.GenerateNextPlanCandidate(ctx, mockUserID, fb)
	if err != nil {
		log.Error("Regenerate plan failed", zap.Error(err))
		respondWithError(w, mapFlowErrorToStatus(err), errorMessage(err), log)
		return
	}

	resp := contract.RegeneratePlanResponse{
		Plan:      models.ConvertToPlanDTO(plan),
		PlanItems: models.ConvertToPlanItemsDTO(items),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, resp, log)
}

func (h *Handler) FinalizePlanSelection(w http.ResponseWriter, r *http.Request) {
	log := h.logger.With(
		zap.String("operation", "finalize_plan"),
		zap.String("user_id", mockUserID.String()),
	)

	log.Info("Finalize plan started")

	dumpID, err := uuid.Parse(chi.URLParam(r, "dump_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "wrong url params", h.logger)
		return
	}

	req, err := parseJSONBody[contract.FinalizePlanRequest](w, r, MaxBodyJSON, log)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body", log)
		return
	}

	planID := req.PlanID
	log = log.With(
		zap.String("dump_id", dumpID.String()),
		zap.String("plan_id", planID.String()),
	)

	ctx := r.Context()
	if err := h.flow.FinalizePlanSelection(ctx, dumpID, planID); err != nil {
		log.Error("Finalize plan failed", zap.Error(err))
		respondWithError(w, mapFlowErrorToStatus(err), errorMessage(err), log)
		return
	}

	resp := contract.FinalizePlanResponse{
		Msg: "plan was saved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encodeResponse(w, resp, log)
}

func respondWithError(w http.ResponseWriter, code int, message string, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := map[string]string{
		"error": message,
	}

	encodeResponse(w, resp, logger)
}

func encodeResponse[T any](w http.ResponseWriter, resp T, logger *zap.Logger) {
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Encode response failed", zap.Error(err))
	}
}

func mapFlowErrorToStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrActiveDumpNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrDumpNotBelongUser):
		return http.StatusForbidden
	case errors.Is(err, service.ErrAnswersAlreadySubmitted):
		return http.StatusConflict
	case errors.Is(err, service.ErrAnalysisNotFound):
		return http.StatusConflict
	case errors.Is(err, service.ErrNoActiveSessionForRegeneration):
		return http.StatusConflict
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func errorMessage(err error) string {
	switch {
	case errors.Is(err, service.ErrActiveDumpNotFound):
		return "active session not found"
	case errors.Is(err, service.ErrDumpNotBelongUser):
		return "dump does not belong to current user"
	case errors.Is(err, service.ErrAnalysisNotFound):
		return "analysis is missing"
	case errors.Is(err, service.ErrAnswersAlreadySubmitted):
		return "answers already submitted"
	case errors.Is(err, service.ErrNoActiveSessionForRegeneration):
		return "no active session available for plan regeneration"
	case errors.Is(err, repository.ErrNotFound):
		return "resource not found"
	default:
		return "internal server error"
	}
}
