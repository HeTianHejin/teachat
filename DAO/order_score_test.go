package dao

import (
	"math"
	"testing"
)

func TestCalculateOrderScoresAttributesByTaskAndContribution(t *testing.T) {
	teamScore, results, err := CalculateOrderScores([]TaskScoreInput{
		{TaskWeight: 30, TaskScore: 90, Participants: []ParticipantScoreInput{{UserId: 1, ResponsibilityWeight: 100, ActualContribution: 100, PerformanceScore: 90, EvidenceScore: 100}}},
		{TaskWeight: 50, TaskScore: 84, Participants: []ParticipantScoreInput{
			{UserId: 2, ResponsibilityWeight: 80, ActualContribution: 100, PerformanceScore: 80, EvidenceScore: 90},
			{UserId: 3, ResponsibilityWeight: 20, ActualContribution: 100, PerformanceScore: 70, EvidenceScore: 80},
		}},
		{TaskWeight: 20, TaskScore: 90, Participants: []ParticipantScoreInput{{UserId: 4, ResponsibilityWeight: 100, ActualContribution: 100, PerformanceScore: 95, EvidenceScore: 100}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(teamScore-85.5) > 0.001 {
		t.Fatalf("team score = %v, want 85.5", teamScore)
	}
	if len(results) != 4 {
		t.Fatalf("got %d member scores, want 4", len(results))
	}
	if math.Abs(results[1].ContributionScore-42) > 0.001 || math.Abs(results[2].ContributionScore-10.5) > 0.001 {
		t.Fatalf("unexpected task contribution split: %#v", results)
	}
}

func TestCalculateOrderScoresUsesActualContribution(t *testing.T) {
	_, results, err := CalculateOrderScores([]TaskScoreInput{{TaskWeight: 100, TaskScore: 80, Participants: []ParticipantScoreInput{
		{UserId: 1, ResponsibilityWeight: 100, ActualContribution: 100, PerformanceScore: 80, EvidenceScore: 80},
		{UserId: 2, ResponsibilityWeight: 100, ActualContribution: 0, PerformanceScore: 100, EvidenceScore: 100},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if results[0].ContributionScore != 80 || results[1].ContributionScore != 0 {
		t.Fatalf("role did not override actual contribution: %#v", results)
	}
}
