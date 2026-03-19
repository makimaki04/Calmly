<div align="center" id="header">
  <h1>☁️ Calmly</h1>
  <p><strong>AI-assisted self-reflection and planning app</strong></p>
</div>

## Overview

Calmly is an AI-assisted self-reflection and planning application.

The current repository is focused on the backend foundation, which powers a guided flow where a user writes down their thoughts, worries, or tasks in free-form text. The system stores this input as a session dump, analyzes it, generates follow-up questions, and then uses the user’s answers to build one or more plan candidates.

If the user is not satisfied with the generated plan, they can provide additional clarification and request another version. Multiple plan candidates can exist within the same session until the user chooses one to save.

A frontend client will be added later to provide the full interactive user experience on top of this backend flow.

## How It Works

A typical session goes through the following steps:

1. The user submits a free-form dump.
2. The system creates a session dump and stores the raw text.
3. The AI analyzes the dump and produces:
   - extracted tasks with category and priority
   - follow-up clarification questions
   - mood and a short supportive quote
4. The user answers the follow-up questions.
5. The system generates an initial structured plan.
6. If needed, the user provides feedback and requests another plan candidate.
7. The user selects one final plan to save.

Once a plan is selected:

- the chosen plan is saved to the user profile
- unsaved plan candidates are removed from that session
- raw session text can later be cleared by TTL for privacy
- the session is considered closed, while the selected result remains the saved outcome

## Current Capabilities

At the current stage, the backend already supports the core guided session flow:

- creating a new dump session from free-form user text
- AI-based dump analysis
- generation of structured follow-up questions
- submission of answers for deeper clarification
- generation of an initial plan candidate
- regeneration of alternative plan candidates based on user feedback
- final selection and saving of one plan
- cleanup of unsaved candidates after finalization

The system already integrates with an LLM provider for real analysis and plan generation.

## Architecture

At the current stage, the project is built as a layered Go backend with:

- repository layer for persistence
- service layer for business logic
- orchestration services for multi-step session flows
- PostgreSQL migrations for schema management
- LLM client layer for analysis and plan generation

The backend is designed around explicit session orchestration and transactional consistency for critical multi-step operations.

## LLM Usage

The AI layer is responsible for two core tasks:

- analyzing the original free-form dump into structured tasks and clarification questions
- generating and regenerating actionable plans based on:
  - the original dump
  - the preliminary analysis
  - the user’s answers
  - the user’s feedback on previous plan candidates

Prompts are structured to keep outputs predictable, machine-readable, and easy to process in the backend.

## Current Scope

This repository currently focuses on backend APIs and domain logic.

The frontend client is planned as the next major step, but the main session and planning workflow is already testable through the backend.

## Current Status

Implemented:

- backend session flow
- real AI-based analysis generation
- initial plan generation
- plan regeneration flow
- final plan selection
- transactional persistence for critical multi-step operations
- unit tests for key handlers, services, repositories, and LLM parsing helpers

Planned next:

- user registration and authentication
- frontend client
- guest session support
- TTL-based cleanup and privacy-oriented retention behavior

## Goal

The goal of the project is to build a clean, reliable, and extensible foundation for a full AI-assisted planning product, starting from a solid backend architecture and gradually expanding into a complete user-facing application.
