package modlicense_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/200sc/modlicense"
)

var anyError = fmt.Errorf("any error")

func TestFromModFilePath(t *testing.T) {
	t.Parallel()
	type testCase struct {
		file          string
		expected      modlicense.ModLicenses
		expectedError error
	}
	tcs := []testCase{
		{
			file: filepath.Join(".", "testdata", "oak_go.mod"),
			expected: modlicense.ModLicenses{
				Licenses: map[modlicense.Dependency]modlicense.License{
					{"dmitri.shuralyov.com/gpu/mtl", "v0.0.0-20201218220906-28db891af037"}:    "BSD-3-Clause",
					{"github.com/BurntSushi/xgb", "v0.0.0-20210121224620-deaf085860bc"}:       "BSD-3-Clause",
					{"github.com/BurntSushi/xgbutil", "v0.0.0-20190907113008-ad855c713046"}:   "WTFPL",
					{"github.com/disintegration/gift", "v1.2.1"}:                              "MIT",
					{"github.com/eaburns/flac", "v0.0.0-20171003200620-9a6fb92396d1"}:         "MIT",
					{"github.com/go-gl/glfw/v3.3/glfw", "v0.0.0-20220320163800-277f93cfa958"}: "BSD-3-Clause",
					{"github.com/golang/freetype", "v0.0.0-20170609003504-e2365dfdc4a0"}:      "unknown",
					{"github.com/hajimehoshi/go-mp3", "v0.3.2"}:                               "Apache-2.0",
					{"github.com/jfreymuth/pulse", "v0.1.0"}:                                  "MIT",
					{"github.com/oakmound/alsa", "v0.0.2"}:                                    "MIT",
					{"github.com/oakmound/libudev", "v0.2.1"}:                                 "MIT",
					{"github.com/oakmound/w32", "v2.1.0+incompatible"}:                        "BSD-3-Clause",
					{"github.com/oov/directsound-go", "v0.0.0-20141101201356-e53e59c700bf"}:   "MIT",
					{"golang.org/x/exp", "v0.0.0-20220414153411-bcd21879b8fd"}:                "BSD-3-Clause",
					{"golang.org/x/image", "v0.0.0-20220321031419-a8550c1d254a"}:              "BSD-3-Clause",
					{"golang.org/x/mobile", "v0.0.0-20220325161704-447654d348e3"}:             "BSD-3-Clause",
					{"golang.org/x/sync", "v0.0.0-20210220032951-036812b2e83c"}:               "BSD-3-Clause",
					{"golang.org/x/sys", "v0.0.0-20220403205710-6acee93ad0eb"}:                "BSD-3-Clause",
					{"github.com/eaburns/bit", "v0.0.0-20131029213740-7bd5cd37375d"}:          "MIT",
				},
			},
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.file, func(t *testing.T) {
			t.Parallel()
			got, gotErr := modlicense.FromModFilePath(tc.file)
			if errors.Is(tc.expectedError, anyError) {
				if gotErr == nil {
					t.Errorf("expected any error, got nil")
				}
			} else {
				if gotErr != nil && !errors.Is(gotErr, tc.expectedError) {
					t.Errorf("expected error %v, got %v", tc.expectedError, gotErr)
				}
			}
			if len(got.Licenses) != len(tc.expected.Licenses) {
				t.Errorf("license count mismatch: got %v expected %v", len(got.Licenses), len(tc.expected.Licenses))
				return
			}
			for k, v := range tc.expected.Licenses {
				v2, ok := got.Licenses[k]
				if !ok {
					t.Errorf("expected license for dependency %v, got nothing", k)
					break
				}
				if v2 != v {
					t.Errorf("expected license for dependency %v: %v, got %v", k, v, v2)
				}
			}
		})
	}
}
