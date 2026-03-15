---
name: readme_documentation
description: Create or update a detailed, well-structured README.md documentation for the project.
---

# README Documentation Generation Skill

## 🎯 Objective

Create or update the project's `README.md` following a standardized, pedagogical, and highly structured format to ensure maximum comprehension for new developers.

## 🗣️ User Interaction Flow

**Always ask the user** before starting the generation/update process:

1. Do you want to **update/create the complete file**?
2. Or do you want to **create/update only a specific topic**?

## 📝 Mode: Complete Documentation

If the user selects the "Complete Documentation" mode:

- If `README.md` **does not exist**, generate it entirely from scratch.
- If `README.md` **already exists**, intelligently update it while carefully preserving relevant existing context.

### Required Index Structure

Ensure the document has a complete index table detailing the topics below. Always use reference files like `cmd/message_publisher/main.go` and `cmd/event_driven_consumer/main.go` internally to build context and realistic instances.

1. **Visão Geral** (Overview)
   - Plugin objective summary.
   - Core features.
   - Architectural patterns and approaches utilized.
   - Core directory and plugin structure.
2. **Bootstrap**
   - How to register components.
   - How to start the plugin successfully.
   - How to gracefully shut down the plugin.
3. **Componentes Principais** (Core Components)
   - **Flow Diagram**: Visualizing relationships between distinct components.
   - **Execution Diagram**: Demonstrating execution workflows.
4. **CQRS**
   - Detailed description of the CQRS pattern behavior in this environment.
   - Practical usage examples.
   - Specific Component & Execution diagrams for CQRS.
5. **Padrões de Publicação** (Publishing Patterns)
   - _Sub-topics_: Comandos (Commands), Queries, Eventos (Events).
   - Provide concrete examples and detailed summaries based on GoDoc regarding the publishing flow.
6. **Padrões de Consumo** (Consumption Patterns)
   - _Sub-topics_: Event-Driven, Polling.
   - Provide a pros vs. cons comparative analysis for both consumption mechanisms.
   - Include operational and conceptual descriptions tied with GoDoc references.
7. **Resiliência** (Resiliency)
   - _Sub-topics_: Retry mechanisms, Dead Letter flows.
   - Embed respective flow diagrams and execution diagrams illustrating resilience behaviors.
8. **Kafka**
   - Implementation configuration details.
   - Deep dive into how the local Kafka driver works.
9. **RabbitMQ**
   - Implementation configuration details.
   - Deep dive into how the local RabbitMQ driver works.

## ✂️ Mode: Specific Topic Insertion

If the user selects the "Specific Topic" mode:

1. **Ask** the user precisely where within the current README index this topical entry should be injected.
2. **Update the index** table to reflect the new topical reference.
3. **Generate the topio content** cleanly at the newly specified position.

## ✍️ Writing Guidelines & Tone

- **Language Constraint**: Always output responses in clear, highly readable **Portuguese**.
- **Target Audience Relevance**: Focus on Developers, particularly aiming to be pedagogical for Junior Developers.
- **Tone Profile**: Didactic, explanatory, and incredibly welcoming.
- **Robust Examples**: Intertwine all explanations with practical code occurrences and configurations.
- **Visualization**: Use diagrams actively (such as Mermaid syntax blocks formatted distinctly for Markdown rendering) where it helps abstractly complex concepts.
