package types

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"strconv"
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

func (pvs *PrivateValidatorState) CompareState(other *PrivateValidatorState) (cmp int, differentSigns bool) {
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
	} else {
		return 1, true
	}

	return 0, pvs.Signature == other.Signature && pvs.SignBytes == other.SignBytes
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
