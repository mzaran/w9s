package dao_test

import (
	"encoding/json"
	"testing"

	"github.com/mzaran/w9s/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWWBoolUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected dao.WWBool
	}{
		{"string true", `"true"`, dao.WWBoolTrue},
		{"string yes", `"yes"`, dao.WWBoolTrue},
		{"string YES", `"YES"`, dao.WWBoolTrue},
		{"string 1", `"1"`, dao.WWBoolTrue},
		{"string True", `"True"`, dao.WWBoolTrue},
		{"string false", `"false"`, dao.WWBoolFalse},
		{"string no", `"no"`, dao.WWBoolFalse},
		{"string NO", `"NO"`, dao.WWBoolFalse},
		{"string 0", `"0"`, dao.WWBoolFalse},
		{"string False", `"False"`, dao.WWBoolFalse},
		{"empty string", `""`, dao.WWBoolFalse},
		{"string UNDEF", `"UNDEF"`, dao.WWBoolUndef},
		{"string undef", `"undef"`, dao.WWBoolUndef},
		{"native true", `true`, dao.WWBoolTrue},
		{"native false", `false`, dao.WWBoolFalse},
		{"string with spaces", `" true "`, dao.WWBoolTrue},
		{"string with spaces false", `" false "`, dao.WWBoolFalse},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b dao.WWBool
			err := json.Unmarshal([]byte(tc.input), &b)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, b)
		})
	}
}

func TestWWBoolUnmarshalUnknownValue(t *testing.T) {
	var b dao.WWBool
	err := json.Unmarshal([]byte(`"something_else"`), &b)
	require.NoError(t, err)
	assert.Equal(t, dao.WWBool("something_else"), b)
}

func TestWWBoolUnmarshalInvalidJSON(t *testing.T) {
	var b dao.WWBool
	err := json.Unmarshal([]byte(`[1,2,3]`), &b)
	require.Error(t, err)
}

func TestWWBoolMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    dao.WWBool
		expected string
	}{
		{"true", dao.WWBoolTrue, `"true"`},
		{"false", dao.WWBoolFalse, `"false"`},
		{"undef", dao.WWBoolUndef, `"UNDEF"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}

func TestWWBoolRoundTrip(t *testing.T) {
	values := []dao.WWBool{dao.WWBoolTrue, dao.WWBoolFalse, dao.WWBoolUndef}

	for _, v := range values {
		t.Run(string(v), func(t *testing.T) {
			data, err := json.Marshal(v)
			require.NoError(t, err)

			var result dao.WWBool
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)
			assert.Equal(t, v, result)
		})
	}
}

func TestWWBoolHelpers(t *testing.T) {
	assert.True(t, dao.WWBoolTrue.IsTrue())
	assert.False(t, dao.WWBoolTrue.IsFalse())
	assert.False(t, dao.WWBoolTrue.IsUnset())

	assert.False(t, dao.WWBoolFalse.IsTrue())
	assert.True(t, dao.WWBoolFalse.IsFalse())
	assert.False(t, dao.WWBoolFalse.IsUnset())

	assert.False(t, dao.WWBoolUndef.IsTrue())
	assert.False(t, dao.WWBoolUndef.IsFalse())
	assert.True(t, dao.WWBoolUndef.IsUnset())

	// Empty string is also considered unset.
	var empty dao.WWBool
	assert.True(t, empty.IsUnset())
}

func TestWWBoolString(t *testing.T) {
	assert.Equal(t, "true", dao.WWBoolTrue.String())
	assert.Equal(t, "false", dao.WWBoolFalse.String())
	assert.Equal(t, "UNDEF", dao.WWBoolUndef.String())
}

func TestWWBoolInStruct(t *testing.T) {
	jsonData := `{
		"discoverable": "yes",
		"asset key": "KEY1",
		"profiles": ["default"],
		"image name": "rocky9",
		"network devices": {
			"eth0": {
				"onbot": "true",
				"ipaddr": "10.0.0.1"
			}
		},
		"ipmi": {
			"write": "false"
		}
	}`

	var node dao.WwNode
	err := json.Unmarshal([]byte(jsonData), &node)
	require.NoError(t, err)

	assert.True(t, node.Discoverable.IsTrue())
	assert.Equal(t, "KEY1", node.AssetKey)
	require.NotNil(t, node.NetDevs["eth0"])
	assert.True(t, node.NetDevs["eth0"].OnBoot.IsTrue())
	require.NotNil(t, node.Ipmi)
	assert.True(t, node.Ipmi.Write.IsFalse())
}

func TestWWBoolNativeJSONBoolInStruct(t *testing.T) {
	// Some API versions might return native JSON booleans.
	jsonData := `{
		"discoverable": true,
		"network devices": {
			"eth0": {
				"onbot": false
			}
		}
	}`

	var node dao.WwNode
	err := json.Unmarshal([]byte(jsonData), &node)
	require.NoError(t, err)

	assert.True(t, node.Discoverable.IsTrue())
	require.NotNil(t, node.NetDevs["eth0"])
	assert.True(t, node.NetDevs["eth0"].OnBoot.IsFalse())
}
