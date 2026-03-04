<div align="center" id="header">
  <h1>☁️ Calmly</h1>
  <p><strong>AI-assisted self-reflection and planning app</strong></p>
</div>

## Overview

**Calmly** is an AI-assisted self-reflection and planning application.

The current repository is focused on the **backend foundation**, which powers a guided flow where a user writes down their thoughts, worries, or tasks in free-form text.  
The system stores this input as a **session dump**, analyzes it, generates follow-up questions, and then uses the user’s answers to build one or more **plan candidates**.

If the user is not satisfied with the generated plan, they can provide additional clarification and request another version.  
Multiple plan candidates can exist within the same session until the user chooses one to save.

A frontend client will be added later to provide the full interactive user experience on top of this backend flow.

## How It Works

Once a plan is selected:

- **The chosen plan is saved** to the user profile
- **Unsaved plan candidates are removed** from that session
- **Raw session text is cleared later by TTL** for privacy

The session itself remains in history, while the selected result is preserved as the final saved plan.

## Architecture

At the current stage, the project is built as a layered **Go backend** with:

- **Repository layer** for persistence
- **Service layer** for business logic
- **PostgreSQL migrations** for schema management
- **Orchestration services** for multi-step session flows

The frontend layer is planned as the next major step after the backend flow is completed.

## Goal

The goal of the project is to build a **clean**, **reliable**, and **extensible** foundation for a full AI-assisted planning product, starting from a solid backend architecture and expanding into a complete user-facing application.
