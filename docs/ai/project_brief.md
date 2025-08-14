# Project Brief

---

## 1. Project Overview

> **Instructions:** Provide a concise, high-level summary of the project. This should be the "elevator pitch" that gives any agent immediate context on the project's purpose and what it aims to achieve.

### 1.1. Project Purpose

> **Instructions:** Describe the "why" behind this project. What problem does it solve? What opportunity does it capture?
>
> *Example: "The purpose of this project is to create a scalable backend service for processing and analyzing real-time IoT sensor data. It will replace the current manual, error-prone system, enabling faster insights and operational efficiency."*

[Describe the project's primary purpose here.]

### 1.2. Goals

> **Instructions:** List the primary, high-level goals of the project. These should be specific, measurable, and aligned with the business requirements.
>
> *Example:*
> *- **Goal 1:** Develop a system capable of ingesting 10,000 data points per second.
> *- **Goal 2:** Reduce data processing time from 24 hours to under 5 minutes.
> *- **Goal 3:** Provide a secure REST API for data retrieval.*

*   **Goal 1:** [Primary objective]
*   **Goal 2:** [Secondary objective]
*   **Goal 3:** [Tertiary objective]

---

## 2. Business Requirements

> **Instructions:** Detail the specific needs the business requires from the final product. This section translates the high-level goals into functional and non-functional requirements.

### 2.1. Functional Requirements

> **Instructions:** List what the system must *do*. Think in terms of features and user actions.
>
> *Example:*
> *- The system must authenticate users via API keys.
> *- The system must allow users to submit data via a `/ingest` endpoint.
> *- The system must provide an endpoint to retrieve aggregated data by date range.*

*   Requirement 1: [Detailed functional requirement]
*   Requirement 2: [Detailed functional requirement]
*   Requirement 3: [Detailed functional requirement]

### 2.2. Non-Functional Requirements

> **Instructions:** List the qualities the system must *have*. This is critical for the Architect agent.
>
> *Example:*
> *- **Performance:** API endpoints must respond in under 200ms.
> *- **Scalability:** The system must be able to scale horizontally to handle traffic spikes.
> *- **Security:** All data in transit and at rest must be encrypted.
> *- **Reliability:** The system must have an uptime of 99.9%.*

*   **Performance:** [Performance targets]
*   **Scalability:** [Scalability requirements]
*   **Security:** [Security standards and requirements]
*   **Reliability:** [Uptime/availability targets]

---

## 3. Success Metrics

> **Instructions:** Define the key performance indicators (KPIs) that will be used to measure the project's success after launch. This helps focus the development effort on what truly matters.
>
> *Example:*
> *- **Metric 1:** Average API response time.
> *- **Metric 2:** System error rate below 0.1%.
> *- **Metric 3:** Monthly infrastructure cost reduction of 15%.*

*   **Metric 1:** [Name of metric and its target value]
*   **Metric 2:** [Name of metric and its target value]
*   **Metric 3:** [Name of metric and its target value]

---

## 4. Key Features (Initial High-Level Scope)

> **Instructions:** List the major features or components that are envisioned for the initial version of the project. This list is not the detailed plan, but rather the high-level "epics" that the Plan agent will break down later.

*   **User Authentication Service:** Handles user registration, login, and API key management.
*   **Data Ingestion API:** A set of endpoints for receiving and validating incoming data.
*   **Asynchronous Processing Engine:** A background worker system for data transformation and analysis.
*   **Data Query API:** A secure API for clients to retrieve processed data.
*   **Admin Dashboard Backend:** APIs to support an administrative interface for monitoring and management.

---

## 5. Out of Scope

> **Instructions:** Clearly state what this project will *not* include. This is crucial for preventing scope creep and ensuring the agents do not build unnecessary functionality.
>
> *Example:*
> *- A customer-facing web user interface.
> *- Real-time push notifications.
> *- Support for third-party identity providers (e.g., Google/Facebook login) in V1.*

*   [Feature or functionality explicitly excluded from this project]
*   [Another feature or functionality explicitly excluded]