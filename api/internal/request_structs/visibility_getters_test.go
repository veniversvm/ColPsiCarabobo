package request_structs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVisibilityGetters_PsiUserUpdateRequestSelf(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		expect *bool
	}{
		{"true", "true", ptrBool(true)},
		{"false", "false", ptrBool(false)},
		{"one", "1", ptrBool(true)},
		{"zero", "0", ptrBool(false)},
		{"empty_returns_nil", "", nil},
		{"garbage_returns_nil", "xyz", nil},
		{"whitespace_true", "  true  ", ptrBool(true)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := PsiUserUpdateRequestSelf{
				ShowContactEmailRaw:                  tc.raw,
				ShowPublicServiceAddressRaw:          tc.raw,
				ShowMunicipalityCaraboboRaw:          tc.raw,
				ShowPhoneCaraboboRaw:                 tc.raw,
				ShowCelPhoneCaraboboRaw:              tc.raw,
				ShowStateOutsideRaw:                  tc.raw,
				ShowMunicipalityOutSideCaraboboRaw:   tc.raw,
				ShowPhoneOutSideCaraboboRaw:          tc.raw,
				ShowCellPhoneOutSideCaraboboRaw:      tc.raw,
				ShowPublicServiceAddressOutSideCaraboboRaw: tc.raw,
				ShowPhoneOutSideVenezuelaRaw:                tc.raw,
				ShowCellPhoneOutSideVenezuelaRaw:            tc.raw,
				ShowPublicServiceAddressOutSideVenezuelaRaw: tc.raw,
				ShowUniversityUndergraduateRaw:             tc.raw,
				ShowGraduateDateRaw:                        tc.raw,
				ShowMentionUndergraduateRaw:                tc.raw,
			}

			// Test all getters return the same result
			require.Equal(t, tc.expect, req.ShowContactEmail())
			require.Equal(t, tc.expect, req.ShowPublicServiceAddress())
			require.Equal(t, tc.expect, req.ShowMunicipalityCarabobo())
			require.Equal(t, tc.expect, req.ShowPhoneCarabobo())
			require.Equal(t, tc.expect, req.ShowCelPhoneCarabobo())
			require.Equal(t, tc.expect, req.ShowStateOutside())
			require.Equal(t, tc.expect, req.ShowMunicipalityOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowPhoneOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowCellPhoneOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowPublicServiceAddressOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowPhoneOutSideVenezuela())
			require.Equal(t, tc.expect, req.ShowCellPhoneOutSideVenezuela())
			require.Equal(t, tc.expect, req.ShowPublicServiceAddressOutSideVenezuela())
			require.Equal(t, tc.expect, req.ShowUniversityUndergraduate())
			require.Equal(t, tc.expect, req.ShowGraduateDate())
			require.Equal(t, tc.expect, req.ShowMentionUndergraduate())
		})
	}
}

func TestVisibilityGetters_UpdatePsiAdminRequest(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		expect *bool
	}{
		{"true", "true", ptrBool(true)},
		{"false", "false", ptrBool(false)},
		{"one", "1", ptrBool(true)},
		{"zero", "0", ptrBool(false)},
		{"empty_returns_nil", "", nil},
		{"garbage_returns_nil", "xyz", nil},
		{"whitespace_true", "  true  ", ptrBool(true)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := UpdatePsiAdminRequest{
				ShowContactEmailRaw:                  tc.raw,
				ShowPublicServiceAddressRaw:          tc.raw,
				ShowMunicipalityCaraboboRaw:          tc.raw,
				ShowPhoneCaraboboRaw:                 tc.raw,
				ShowCelPhoneCaraboboRaw:              tc.raw,
				ShowStateOutsideRaw:                  tc.raw,
				ShowMunicipalityOutSideCaraboboRaw:   tc.raw,
				ShowPhoneOutSideCaraboboRaw:          tc.raw,
				ShowCellPhoneOutSideCaraboboRaw:      tc.raw,
				ShowPublicServiceAddressOutSideCaraboboRaw: tc.raw,
				ShowPhoneOutSideVenezuelaRaw:                tc.raw,
				ShowCellPhoneOutSideVenezuelaRaw:            tc.raw,
				ShowPublicServiceAddressOutSideVenezuelaRaw: tc.raw,
				ShowUniversityUndergraduateRaw:             tc.raw,
				ShowGraduateDateRaw:                        tc.raw,
				ShowMentionUndergraduateRaw:                tc.raw,
			}

			require.Equal(t, tc.expect, req.ShowContactEmail())
			require.Equal(t, tc.expect, req.ShowPublicServiceAddress())
			require.Equal(t, tc.expect, req.ShowMunicipalityCarabobo())
			require.Equal(t, tc.expect, req.ShowPhoneCarabobo())
			require.Equal(t, tc.expect, req.ShowCelPhoneCarabobo())
			require.Equal(t, tc.expect, req.ShowStateOutside())
			require.Equal(t, tc.expect, req.ShowMunicipalityOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowPhoneOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowCellPhoneOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowPublicServiceAddressOutSideCarabobo())
			require.Equal(t, tc.expect, req.ShowPhoneOutSideVenezuela())
			require.Equal(t, tc.expect, req.ShowCellPhoneOutSideVenezuela())
			require.Equal(t, tc.expect, req.ShowPublicServiceAddressOutSideVenezuela())
			require.Equal(t, tc.expect, req.ShowUniversityUndergraduate())
			require.Equal(t, tc.expect, req.ShowGraduateDate())
			require.Equal(t, tc.expect, req.ShowMentionUndergraduate())
		})
	}
}

func TestVisibilityGetters_PsiUserUpdateRequestSelf_MixedFields(t *testing.T) {
	req := PsiUserUpdateRequestSelf{
		ShowContactEmailRaw:         "true",
		ShowPublicServiceAddressRaw: "false",
		ShowMunicipalityCaraboboRaw: "1",
		ShowPhoneCaraboboRaw:        "0",
		ShowCelPhoneCaraboboRaw:     "",
	}

	require.Equal(t, ptrBool(true), req.ShowContactEmail())
	require.Equal(t, ptrBool(false), req.ShowPublicServiceAddress())
	require.Equal(t, ptrBool(true), req.ShowMunicipalityCarabobo())
	require.Equal(t, ptrBool(false), req.ShowPhoneCarabobo())
	require.Nil(t, req.ShowCelPhoneCarabobo())
}

func TestVisibilityGetters_UpdatePsiAdminRequest_MixedFields(t *testing.T) {
	req := UpdatePsiAdminRequest{
		ShowContactEmailRaw:         "true",
		ShowPublicServiceAddressRaw: "false",
		ShowMunicipalityCaraboboRaw: "1",
		ShowPhoneCaraboboRaw:        "0",
		ShowCelPhoneCaraboboRaw:     "",
	}

	require.Equal(t, ptrBool(true), req.ShowContactEmail())
	require.Equal(t, ptrBool(false), req.ShowPublicServiceAddress())
	require.Equal(t, ptrBool(true), req.ShowMunicipalityCarabobo())
	require.Equal(t, ptrBool(false), req.ShowPhoneCarabobo())
	require.Nil(t, req.ShowCelPhoneCarabobo())
}

func ptrBool(b bool) *bool {
	return &b
}
