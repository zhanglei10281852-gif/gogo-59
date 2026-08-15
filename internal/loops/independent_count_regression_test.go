package loops

import (
	"testing"

	"CaveLoop/internal/config"
	"CaveLoop/internal/geom"
	"CaveLoop/internal/model"
	"CaveLoop/internal/network"
	"CaveLoop/internal/reduce"
	"CaveLoop/internal/traverse"
)

// TestIndependentCountCoversDisconnectedComponents checks the reported size of
// the loop basis on surveys that fall apart into several disconnected parts. The
// count must always agree with the number of circuits that are actually built.
func TestIndependentCountCoversDisconnectedComponents(t *testing.T) {
	origin := geom.Vector{}
	twoLoops := reduce.Result{
		Cave: "Split Cave",
		Stations: []reduce.Station{
			{Name: "A", Flags: []string{model.FlagFixed}, Fixed: &origin},
			{Name: "B"}, {Name: "C"},
			{Name: "P"}, {Name: "Q"}, {Name: "R"},
		},
		Shots: []reduce.Shot{
			splitLeg("S1", "A", "B", geom.Vector{East: 10}),
			splitLeg("S2", "B", "C", geom.Vector{North: 10}),
			splitLeg("S3", "A", "C", geom.Vector{East: 10, North: 10}),
			splitLeg("S4", "P", "Q", geom.Vector{East: 8}),
			splitLeg("S5", "Q", "R", geom.Vector{North: 8}),
			splitLeg("S6", "P", "R", geom.Vector{East: 8, North: 8}),
		},
	}
	detected := detectSplit(t, twoLoops)
	if len(detected.Loops) != 2 {
		t.Fatalf("two disconnected circuits produced %d loops", len(detected.Loops))
	}
	if detected.IndependentCount != 2 {
		t.Fatalf("independent count is %d, want 2", detected.IndependentCount)
	}
	if detected.IndependentCount != len(detected.Loops) {
		t.Fatalf("independent count %d disagrees with the %d circuits that were built",
			detected.IndependentCount, len(detected.Loops))
	}
	components := map[int]bool{}
	for _, loop := range detected.Loops {
		components[loop.Component] = true
	}
	if len(components) != 2 {
		t.Fatalf("the circuits report components %v, want one loop per component", components)
	}

	oneLoopOneChain := reduce.Result{
		Cave: "Split Cave",
		Stations: []reduce.Station{
			{Name: "A", Flags: []string{model.FlagFixed}, Fixed: &origin},
			{Name: "B"}, {Name: "C"}, {Name: "P"}, {Name: "Q"},
		},
		Shots: []reduce.Shot{
			splitLeg("S1", "A", "B", geom.Vector{East: 10}),
			splitLeg("S2", "B", "C", geom.Vector{North: 10}),
			splitLeg("S3", "A", "C", geom.Vector{East: 10, North: 10}),
			splitLeg("S4", "P", "Q", geom.Vector{East: 8}),
		},
	}
	mixed := detectSplit(t, oneLoopOneChain)
	if len(mixed.Loops) != 1 {
		t.Fatalf("a loop plus a dangling branch produced %d loops", len(mixed.Loops))
	}
	if mixed.IndependentCount != 1 {
		t.Fatalf("independent count is %d, want 1", mixed.IndependentCount)
	}
}

// splitLeg builds a reduced leg for the regression fixtures.
func splitLeg(shotID, from, to string, vector geom.Vector) reduce.Shot {
	return reduce.Shot{
		TripID:         "T1",
		ShotID:         shotID,
		From:           from,
		To:             to,
		DistanceMeters: vector.Length(),
		Vector:         vector,
	}
}

// detectSplit runs the graph, layout and loop stages over a reduction result.
func detectSplit(t *testing.T, result reduce.Result) Result {
	t.Helper()
	graph := network.Build(result)
	analysis := network.Analyze(graph)
	layout, err := traverse.Compute(graph, analysis, nil)
	if err != nil {
		t.Fatalf("Compute returned %v", err)
	}
	detected, err := Detect(graph, analysis, layout, config.Default().Tolerances, nil)
	if err != nil {
		t.Fatalf("Detect returned %v", err)
	}
	return detected
}
