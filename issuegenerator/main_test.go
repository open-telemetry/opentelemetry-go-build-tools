package main

import (
	"testing"

	cdowners "github.com/hmarr/codeowners"
	"github.com/joshdk/go-junit"
	"github.com/stretchr/testify/require"
)

func TestIngestArtifacts(t *testing.T) {
	rg := newReportGenerator()
	rg.ingestArtifacts("./testdata/junit", "./testdata/codeowners/CODEOWNERS_good")

	expectedTestResults := map[string]junit.Suite{
		"package1": junit.Suite{
			Name:       "package1",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				junit.Test{
					Name:      "TestFailure",
					Classname: "package1",
					Duration:  0,
					Status:    "failed",
					Message:   "Failed",
					Error: junit.Error{
						Message: "Failed",
						Type:    "",
						Body:    "=== RUN   TestFailure\n--- FAIL: TestFailure (0.00s)\n",
					},
					Properties: map[string]string{"classname": "package1", "name": "TestFailure", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  "",
				},
				junit.Test{
					Name:       "TestSucess",
					Classname:  "package1",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package1", "name": "TestSucess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  ""},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    2,
				Passed:   1,
				Skipped:  0,
				Failed:   1,
				Error:    0,
				Duration: 0,
			},
		}, "package2": junit.Suite{
			Name:       "package2",
			Package:    "",
			Properties: map[string]string{"go.version": "go1.23.1 darwin/arm64"},
			Tests: []junit.Test{
				junit.Test{
					Name:      "TestFailure",
					Classname: "package2",
					Duration:  0,
					Status:    "failed",
					Message:   "Failed",
					Error: junit.Error{
						Message: "Failed",
						Type:    "",
						Body:    "=== RUN   TestFailure\n--- FAIL: TestFailure (0.00s)\n",
					}, Properties: map[string]string{"classname": "package2", "name": "TestFailure", "time": "0.000000"},
					SystemOut: "",
					SystemErr: ""},
				junit.Test{
					Name:       "TestSucess",
					Classname:  "package2",
					Duration:   0,
					Status:     "passed",
					Message:    "",
					Properties: map[string]string{"classname": "package2", "name": "TestSucess", "time": "0.000000"},
					SystemOut:  "",
					SystemErr:  ""},
			},
			SystemOut: "",
			SystemErr: "",
			Totals: junit.Totals{
				Tests:    2,
				Passed:   1,
				Skipped:  0,
				Failed:   1,
				Error:    0,
				Duration: 0},
		},
	}
	require.Equal(t, expectedTestResults, rg.testSuites)

	expectedCodeowners := cdowners.Ruleset{
		{LineNumber: 1, Owners: []cdowners.Owner{{Value: "User1", Type: "username"}}},
		{LineNumber: 2, Owners: []cdowners.Owner{{Value: "User2", Type: "username"}}},
	}
	// We can't match the whole struct because the regex pattern is not exported.
	require.Equal(t, expectedCodeowners[0].Owners, rg.codeowners[0].Owners)
	require.Equal(t, expectedCodeowners[0].LineNumber, rg.codeowners[0].LineNumber)
	require.Equal(t, expectedCodeowners[1].Owners, rg.codeowners[1].Owners)
	require.Equal(t, expectedCodeowners[1].LineNumber, rg.codeowners[1].LineNumber)
}

// func TestProcessTestResults(t *testing.T) {
// 	testCases := []struct {
// 		name       string
// 		testSuite  map[string]junit.Suite
// 		codeowners cdowners.Ruleset

// 		expectedReports []report
// 	}{
// 		{
// 			name: "Default codeowners",
// 			testSuite: map[string]junit.Suite{
// 				"package1": junit.Suite{
// 					Name:  "package1",
// 					Tests: []junit.Test{{Name: "TestFailure", Status: junit.StatusFailed}},
// 					Totals: junit.Totals{
// 						Failed: 1,
// 					},
// 				},
// 				"package2": junit.Suite{
// 					Name:  "package2",
// 					Tests: []junit.Test{{Name: "TestFailure", Status: junit.StatusFailed}},
// 					Totals: junit.Totals{
// 						Failed: 1,
// 					},
// 				},
// 			},
// 			codeowners: func() cdowners.Ruleset {
// 				f, _ := os.Open("./testdata/codeowners/CODEOWNERS_good")
// 				rs, _ := cdowners.ParseFile(f)
// 				return rs
// 			}(),

// 			expectedReports: []report{
// 				{module: "package1", codeOwners: "@User1, @User2", failedTests: []string{"TestFailure"}},
// 				{module: "package2", codeOwners: "@User1, @User2", failedTests: []string{"TestFailure"}},
// 			},
// 		},
// 		{
// 			name: "Overlapping codeowners",
// 			testSuite: map[string]junit.Suite{
// 				"package1": junit.Suite{
// 					Name:  "package1",
// 					Tests: []junit.Test{{Name: "TestFailure", Status: junit.StatusFailed}},
// 					Totals: junit.Totals{
// 						Failed: 1,
// 					},
// 				},
// 			},
// 			codeowners: func() cdowners.Ruleset {
// 				f, _ := os.Open("./testdata/codeowners/CODEOWNERS_good")
// 				rs, _ := cdowners.ParseFile(f)
// 				return rs
// 			}(),
// 			expectedReports: []report{
// 				{module: "package1", codeOwners: "@User2", failedTests: []string{"TestFailure"}},
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			rg := newReportGenerator()
// 			rg.testSuites = tc.testSuite
// 			rg.codeowners = tc.codeowners
// 			rg.processTestResults()

// 			require.Equal(t, tc.expectedReports, rg.reports)
// 		})
// 	}
// }
