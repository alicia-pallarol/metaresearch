-- Seed the twelve research areas from the framework table.
-- Idempotent: re-running updates the definition/name/group but preserves ids.

INSERT INTO research_areas (problem_group, tag, name, definition, sort_order) VALUES
('Specification and objective', 'EA', 'Empirical Alignment',
 'Training models using RLHF, fine-tuning, and feedback signals to align with human preferences; audits for capability spillovers in training pipelines.', 1),
('Interpretability and honesty', 'IN', 'Interpretability',
 'Mechanistic and developmental approaches to auditing AI internals, detecting deception or misalignment; moderation protocols to redact circuit insights that could improve capabilities.', 2),
('Interpretability and honesty', 'SA', 'Safety Assurance',
 'Transparency tools, auditing, red teaming, and chain-of-thought monitoring for empirical safety verification; emphasizes aggregate findings rather than exploitable methodologies.', 3),
('Measurement and evaluation', 'CE', 'Capabilities Evaluation',
 'Scientific study of AI capabilities including evaluations of emergence and scaling behavior; benchmarks may be redacted in controlled sharing to avoid informing capability roadmaps.', 4),
('Measurement and evaluation', 'SM', 'Safety Metaresearch',
 'Research on the impact of safety work itself, including CSP frameworks for forecasting spillovers; prioritized for moderation to support safety without generating capability roadmaps.', 5),
('Oversight and control', 'SO', 'Scalable Oversight',
 'Methods such as debate, amplification, and recursive oversight for supervising superhuman AI; contributions may be redacted when they could support capability amplification.', 6),
('Oversight and control', 'DS', 'Deployment Safety',
 'Operational protocols for safe AI rollout including corrigibility and shutdown mechanisms; moderated to exclude techniques reusable for unrestricted agent scaling.', 7),
('Capability and agency dynamics', 'AR', 'Adversarial Robustness',
 'Methods for hardening models against adversarial attacks and distributional shifts; redaction guidelines to prevent robustness tools from enhancing general performance evaluations.', 8),
('Multi-agent and societal', 'ST', 'Socio-Technical Safety',
 'Integration of social, institutional, and technical factors for holistic alignment including governance interfaces; redaction policies may apply to dual-use socio-technical models.', 9),
('Governance and political economy', 'TG', 'Technical AI Governance',
 'Technical tools for governance including verification and policy enforcement mechanisms; contribution protocols prevent formal methods from enabling unchecked scaling.', 10),
('Threat models and field-level', 'AF', 'Alignment Foundations',
 'Theoretical work on alignment problems, threat modeling, agent foundations, and formal verification; strict moderation to avoid foundational insights accelerating power-seeking behaviors.', 11),
('Threat models and field-level', 'SH', 'Security Hardening',
 'Defenses against misuse, jailbreaks, and security vulnerabilities; may require classification where prevention techniques could serve as evasion tools for capabilities research.', 12)
ON CONFLICT (tag) DO UPDATE SET
    problem_group = EXCLUDED.problem_group,
    name          = EXCLUDED.name,
    definition    = EXCLUDED.definition,
    sort_order    = EXCLUDED.sort_order;
