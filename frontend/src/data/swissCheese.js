// Swiss cheese defense-model mapping, transcribed from `swiss_cheese_research.md`.
//
// This content was drafted with AI assistance and has NOT been peer-reviewed.
// The UI must display that disclaimer wherever this data is shown.
export const AI_DISCLAIMER =
  'The conceptual mapping on this page was drafted with AI assistance and has not yet been peer-reviewed.'

export const LAYERS = [
  {
    id: 1,
    title: 'Alignment by construction',
    stage: 'Before / during training',
    hazard: 'Models that learn the wrong objectives in the first place.',
    tags: ['EA', 'AF'],
    holes: ['Reward misspecification', 'Goal misgeneralization', 'Theory–practice gap'],
  },
  {
    id: 2,
    title: 'Transparency and verification',
    stage: 'Understanding what was built',
    hazard: 'Catching misalignment that slipped through training.',
    tags: ['IN', 'SA', 'CE'],
    holes: [
      'Unfaithful explanations',
      'Evals that miss emergent capabilities',
      'Deceptive alignment evading audits',
    ],
  },
  {
    id: 3,
    title: 'Oversight and control',
    stage: 'Supervising operation',
    hazard: 'Keeping humans meaningfully in control of very capable systems.',
    tags: ['SO', 'AR'],
    holes: ['Oversight that does not scale past human evaluators', 'Unknown attack surfaces'],
  },
  {
    id: 4,
    title: 'Deployment and security',
    stage: 'Operational safeguards',
    hazard: 'Containing failures and misuse in the real world.',
    tags: ['DS', 'SH'],
    holes: ['Jailbreaks', 'Insider threat', 'Safeguards removed under competitive pressure'],
  },
  {
    id: 5,
    title: 'Societal and institutional',
    stage: 'The outermost layer',
    hazard: 'Governance, norms, and institutions as the last line of defense.',
    tags: ['ST', 'TG', 'SM'],
    holes: ['Regulatory lag', 'Race dynamics', 'Jurisdictional gaps'],
    note: 'SM (Safety Metaresearch) is cross-cutting: it also monitors hole alignment across all layers.',
  },
]

export const FURTHER_READING = [
  'Reason, J. (1990). Human Error. Cambridge University Press.',
  'Reason, J. (2000). "Human error: models and management." BMJ 320:768–770.',
  'Hendrycks, D. et al. (2023). "An Overview of Catastrophic AI Risks."',
  'Defense-in-depth discussions in frontier-lab safety frameworks.',
]
