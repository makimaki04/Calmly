<div align="center" id="header">
    <h1>☁️ Calmly</h2>
</div>
# Calmly
Calmly is a backend service for a guided self-reflection and planning flow.

A user starts a session by writing down their thoughts, worries, or tasks in free-form text.  
The system stores this input as a session dump, analyzes it, generates follow-up questions, and then uses the user’s answers to build one or more plan candidates.

If the user is not satisfied with the generated plan, they can provide additional clarification and request another version.  
Multiple plan candidates can exist within the same session until the user chooses one to save.

Once a plan is selected:
- the chosen plan is saved to the user profile,
- unsaved plan candidates from that session are removed,
- and the raw session text is cleared later by TTL for privacy.

The project is built as a layered Go backend with:
- repository layer for persistence,
- service layer for business logic,
- migration-based PostgreSQL schema,
- and orchestration services for multi-step session flows.

The main goal of the project is to provide a clean and reliable backend foundation for an AI-assisted planning experience.
