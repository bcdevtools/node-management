package validation

import "regexp"

var regexPeerPlus = regexp.MustCompile(`^[a-f\d]{40}@(([^:]+)|(\[[a-f\d]*(:+[a-f\d]+)+])):\d{1,5}(,[a-f\d]{40}@(([^:]+)|(\[[a-f\d]*(:+[a-f\d]+)+])):\d{1,5})*$`)

func IsValidPeer(peer string) bool {
	return regexPeerPlus.MatchString(peer)
}
