package library

import (
	"encoding/json"
	"testing"
)

// Which instruction set is on is this computer's own choice (#7327,
// MaxMaoSu): a sync that applies a bundle shares the sets and their texts,
// but keeps the set this computer has on. It takes the bundle's only when
// nothing is chosen here yet (a fresh computer joining the folder), or the
// set it uses is no longer among the bundle's (removed on another
// computer).
func TestPutKeepsThisComputersInstructionSet(t *testing.T) {
	// the sender: two sets, "work" on
	h1 := sandbox(t)
	setUpLibrary(t, h1)
	sent, err := Collect()
	if err != nil {
		t.Fatal(err)
	}
	carried := mustJSON(t, sent)

	// this computer is on "default" while the sets are shared: the sync
	// brings the bundle's texts and leaves this computer's choice alone
	sandbox(t)
	home, work := "Home rules.", "Company rules."
	ok(t)(SaveInstructions(InstructionsChange{Shared: &home, Agents: []string{"claude", "codex"}}))
	ok(t)(SaveInstructions(InstructionsChange{Create: &InstrSet{ID: "work", Name: "Work"}, Texts: map[string]*string{"work": &work}}))
	var in Bundle
	json.Unmarshal([]byte(carried), &in)
	ok(t)(Put(&in))
	v, err := ReadInstructions()
	if err != nil {
		t.Fatal(err)
	}
	on, taken := "", ""
	for _, s := range v.Sets {
		if s.Active {
			on = s.ID
		}
		if s.ID == "work" {
			taken = s.Text
		}
	}
	if on != "" && on != defaultSet {
		t.Errorf("this computer's set was flipped by the sync: %q is on, want the default it chose", on)
	}
	if taken != "Company rules." {
		t.Errorf("the bundle's text for \"work\" didn't arrive: %q", taken)
	}

	// the set this computer uses is no longer in the bundle: it moves to
	// the bundle's
	sandbox(t)
	solo := "Just this one."
	ok(t)(SaveInstructions(InstructionsChange{Shared: &home, Agents: []string{"claude", "codex"}}))
	ok(t)(SaveInstructions(InstructionsChange{Create: &InstrSet{ID: "solo", Name: "Solo"}, Texts: map[string]*string{"solo": &solo}, Activate: "solo"}))
	var in2 Bundle
	json.Unmarshal([]byte(carried), &in2)
	ok(t)(Put(&in2))
	v, err = ReadInstructions()
	if err != nil {
		t.Fatal(err)
	}
	on = ""
	for _, s := range v.Sets {
		if s.Active {
			on = s.ID
		}
	}
	if on != "work" {
		t.Errorf("a set the bundle no longer carries left this computer on it: %q is on, want the bundle's \"work\"", on)
	}

	// a fresh computer joins on the sender's set
	sandbox(t)
	var in3 Bundle
	json.Unmarshal([]byte(carried), &in3)
	ok(t)(Put(&in3))
	v, err = ReadInstructions()
	if err != nil {
		t.Fatal(err)
	}
	on = ""
	for _, s := range v.Sets {
		if s.Active {
			on = s.ID
		}
	}
	if on != "work" {
		t.Errorf("a fresh computer didn't join on the sender's set: %q is on, want \"work\"", on)
	}
}
