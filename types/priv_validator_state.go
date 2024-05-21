package types

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

type PrivateValidatorState struct {
	originBz  []byte
	Height    string `json:"height"`
	Round     int    `json:"round"`
	Step      int    `json:"step"`
	Signature string `json:"signature,omitempty"`
	SignBytes string `json:"signbytes,omitempty"`
}

func NewEmptyPrivateValidatorState() PrivateValidatorState {
	return PrivateValidatorState{
		Height:    "0",
		Round:     0,
		Step:      0,
		Signature: "",
		SignBytes: "",
	}
}

func (pvs PrivateValidatorState) IsEmpty() bool {
	return pvs.Height == "0" &&
		pvs.Round == 0 &&
		pvs.Step == 0 &&
		pvs.Signature == "" &&
		pvs.SignBytes == ""
}

func (pvs PrivateValidatorState) Equals(other PrivateValidatorState) bool {
	return pvs.Height == other.Height &&
		pvs.Round == other.Round &&
		pvs.Step == other.Step &&
		pvs.Signature == other.Signature &&
		pvs.SignBytes == other.SignBytes
}

func (pvs PrivateValidatorState) CompareState(other PrivateValidatorState) (cmp int, differentSigns bool) {
	var heightPvs, heightOther int64
	var err error
	if pvs.Height != "0" {
		heightPvs, err = strconv.ParseInt(pvs.Height, 10, 64)
		if err != nil {
			panic(errors.Wrap(err, "failed to parse private validator state height"))
		}
	}
	if other.Height != "0" {
		heightOther, err = strconv.ParseInt(other.Height, 10, 64)
		if err != nil {
			panic(errors.Wrap(err, "failed to parse other private validator state height"))
		}
	}

	if heightPvs < heightOther {
		return -1, true
	} else if heightPvs > heightOther {
		return 1, true
	}

	if pvs.Round < other.Round {
		return -1, true
	} else if pvs.Round > other.Round {
		return 1, true
	}

	if pvs.Step < other.Step {
		return -1, true
	} else if pvs.Step > other.Step {
		return 1, true
	}

	sameSign := pvs.Signature == other.Signature && pvs.SignBytes == other.SignBytes
	return 0, !sameSign
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
	pvs.originBz = bz
	return nil
}

func (pvs *PrivateValidatorState) SaveToJSONFile(filePath string) error {
	jsonStr := pvs.Json()
	err := os.WriteFile(filePath, []byte(jsonStr), 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}
	return nil
}

func (pvs PrivateValidatorState) Json() string {
	if len(pvs.originBz) > 0 {
		return string(pvs.originBz)
	}
	bz, err := json.MarshalIndent(pvs, "", "  ")
	if err != nil {
		panic(errors.Wrap(err, "failed to marshal JSON"))
	}
	return string(bz)
}
