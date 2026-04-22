# 🌿 Calathea

> Agentic SDLC governance platform with living RFCs, reproducible workflows, and AI-assisted development surfaces.

[![Build Status](https://img.shields.io/github/actions/workflow/status/<org>/<repo>/ci.yml?branch=main)](https://github.com/<org>/<repo>/actions)
[![License](https://img.shields.io/github/license/<org>/<repo>)](./LICENSE)
[![Last Commit](https://img.shields.io/github/last-commit/<org>/<repo>)](https://github.com/<org>/<repo>/commits/main)
[![Issues](https://img.shields.io/github/issues/<org>/<repo>)](https://github.com/<org>/<repo>/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/<org>/<repo>)](https://github.com/<org>/<repo>/pulls)
[![Repo Size](https://img.shields.io/github/repo-size/<org>/<repo>)](https://github.com/<org>/<repo>)
[![Stars](https://img.shields.io/github/stars/<org>/<repo>?style=social)](https://github.com/<org>/<repo>)

---

## 🧭 Overview

Calathea is a governance-first development system designed for:

- **Agentic workflows** (AI + human collaboration)
- **Living RFC-driven architecture**
- **Reproducible execution across environments**
- **Secure, policy-aware SDLC**

It aligns planning, implementation, and review into a continuous feedback loop rather than discrete phases.

---

## 🧱 Core Concepts

| Concept | Description |
|--------|-------------|
| **Living RFCs** | Source of truth for architecture and decisions |
| **Paved Road** | Approved tools, patterns, and workflows |
| **Reproducible Execution** | Deterministic builds, runs, and outputs |
| **Governance Loop** | Planning → Implementation → Feedback → Refinement |
| **Agent Surfaces** | AI-assisted components with defined boundaries |

---

## 🏗 Architecture (High-Level)

```mermaid
flowchart LR
    RFCs[Living RFCs] --> Planning
    Planning --> Tasks
    Tasks --> Implementation
    Implementation --> Review
    Review --> Feedback
    Feedback --> RFCs

    subgraph Runtime
        Agents
        Services
        CLI
    end

    Tasks --> Runtime
