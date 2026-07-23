// Package atlas carries the identifiers the API must recognise.
//
// The curated map itself is static and shipped with the frontend; the backend only
// needs the id sets so it can reject feedback for things that do not exist.
// Generated from data/atlas.json by scripts/gen_agenda_ids.py. Do not edit by hand;
// ids_test.go fails loudly if this file drifts from the dataset.
package atlas

// Iteration0AgendaIDs are the 58 agenda ids of the curated map.
var Iteration0AgendaIDs = []string{
	"EA1", "EA2", "EA3", "EA4", "EA5", "EA6",
	"IN1", "IN2", "IN3", "IN4", "IN5", "IN6",
	"IN7", "AR1", "AR2", "AR3", "DS1", "DS2",
	"DS3", "DS4", "SO1", "SO2", "SO3", "SO4",
	"SO5", "SO6", "AF1", "AF2", "AF3", "AF4",
	"AF5", "AF6", "AF7", "AF8", "CE1", "CE2",
	"CE3", "CE4", "CE5", "SA1", "SA2", "SA3",
	"SA4", "SA5", "SA6", "SH1", "SH2", "SH3",
	"ST1", "ST2", "ST3", "SM1", "SM2", "SM3",
	"SM4", "TG1", "TG2", "TG3",
}

// Iteration0AreaTags are the 12 research-area tags.
var Iteration0AreaTags = []string{
	"EA", "IN", "AR", "DS", "SO", "AF",
	"CE", "SA", "SH", "ST", "SM", "TG",
}

// Iteration0ProblemIDs are the 12 problem ids.
var Iteration0ProblemIDs = []string{
	"P1", "P2", "P3", "P4", "P5", "P6",
	"P7", "P8", "P9", "P10", "P11", "P12",
}

// Iteration0Tiers are the 6 maturity tiers of legend.tier_order,
// strongest evidence first. A maturity vote must name one of these.
var Iteration0Tiers = []string{
	"Robust (small scale)",
	"Strong existence proof",
	"Early / partial",
	"Untested",
	"Never demonstrated",
	"Contested",
}

// agendaArea maps each agenda to the research area it belongs to, so a vote
// on a subarea can be checked against the area it was cast under.
var agendaArea = map[string]string{
	"EA1": "EA",
	"EA2": "EA",
	"EA3": "EA",
	"EA4": "EA",
	"EA5": "EA",
	"EA6": "EA",
	"IN1": "IN",
	"IN2": "IN",
	"IN3": "IN",
	"IN4": "IN",
	"IN5": "IN",
	"IN6": "IN",
	"IN7": "IN",
	"AR1": "AR",
	"AR2": "AR",
	"AR3": "AR",
	"DS1": "DS",
	"DS2": "DS",
	"DS3": "DS",
	"DS4": "DS",
	"SO1": "SO",
	"SO2": "SO",
	"SO3": "SO",
	"SO4": "SO",
	"SO5": "SO",
	"SO6": "SO",
	"AF1": "AF",
	"AF2": "AF",
	"AF3": "AF",
	"AF4": "AF",
	"AF5": "AF",
	"AF6": "AF",
	"AF7": "AF",
	"AF8": "AF",
	"CE1": "CE",
	"CE2": "CE",
	"CE3": "CE",
	"CE4": "CE",
	"CE5": "CE",
	"SA1": "SA",
	"SA2": "SA",
	"SA3": "SA",
	"SA4": "SA",
	"SA5": "SA",
	"SA6": "SA",
	"SH1": "SH",
	"SH2": "SH",
	"SH3": "SH",
	"ST1": "ST",
	"ST2": "ST",
	"ST3": "ST",
	"SM1": "SM",
	"SM2": "SM",
	"SM3": "SM",
	"SM4": "SM",
	"TG1": "TG",
	"TG2": "TG",
	"TG3": "TG",
}

// AreaOfAgenda returns the area tag an agenda belongs to.
func AreaOfAgenda(id string) (string, bool) {
	tag, ok := agendaArea[id]
	return tag, ok
}

var agendaIDSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Iteration0AgendaIDs))
	for _, id := range Iteration0AgendaIDs {
		m[id] = struct{}{}
	}
	return m
}()

// KnownAgendaID reports whether id is one of the curated agendas.
func KnownAgendaID(id string) bool {
	_, ok := agendaIDSet[id]
	return ok
}

var areaTagSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Iteration0AreaTags))
	for _, id := range Iteration0AreaTags {
		m[id] = struct{}{}
	}
	return m
}()

// KnownAreaTag reports whether id is one of the research-area tags.
func KnownAreaTag(id string) bool {
	_, ok := areaTagSet[id]
	return ok
}

var problemIDSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Iteration0ProblemIDs))
	for _, id := range Iteration0ProblemIDs {
		m[id] = struct{}{}
	}
	return m
}()

// KnownProblemID reports whether id is one of the problem ids.
func KnownProblemID(id string) bool {
	_, ok := problemIDSet[id]
	return ok
}

var tierSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(Iteration0Tiers))
	for _, id := range Iteration0Tiers {
		m[id] = struct{}{}
	}
	return m
}()

// KnownTier reports whether id is one of the maturity tiers.
func KnownTier(id string) bool {
	_, ok := tierSet[id]
	return ok
}
