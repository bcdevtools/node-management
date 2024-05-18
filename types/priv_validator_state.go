package types

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

type PrivateValidatorState struct {
	Height    string `json:"height"`
	Round     int    `json:"round"`
	Step      int    `json:"step"`
	Signature string `json:"signature"`
	SignBytes string `json:"signbytes"`
}

func (pvs PrivateValidatorState) IsEmpty() bool {
	return pvs.Height == "0" &&
		pvs.Round == 0 &&
		pvs.Step == 0 &&
		pvs.Signature == "" &&
		pvs.SignBytes == ""
}

func (pvs *PrivateValidatorState) Equals(other *PrivateValidatorState) bool {
	return pvs.Height == other.Height &&
		pvs.Round == other.Round &&
		pvs.Step == other.Step &&
		pvs.Signature == other.Signature &&
		pvs.SignBytes == other.SignBytes
}

func (pvs *PrivateValidatorState) LoadFromJSONFile(filePath string) error {
	bz, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}
	return pvs.LoadFromJSON(bz)
}

func (pvs *PrivateValidatorState) LoadFromJSON(bz []byte) error {
	err := json.Unmarshal(bz, pvs)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal JSON")
	}
	return nil
}
