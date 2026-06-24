package store

import (
	"crypto/rand"
	"fmt"
	"time"
)

// crockford is the lowercase Crockford base32 alphabet (no i, l, o, u). It is strictly
// ascending in ASCII, so the lexical order of an encoded fixed-width number matches its
// numeric order — that is what makes time-ordered ids sort chronologically as plain strings.
const crockford = "0123456789abcdefghjkmnpqrstvwxyz"

// idTimeChars is the width of the encoded millisecond timestamp (10 chars × 5 bits = 50 bits,
// comfortably covering the 48-bit UnixMilli range for the next ~8900 years). idRandChars is
// the random tie-break tail that distinguishes ids minted within the same millisecond.
const (
	idTimeChars = 10
	idRandChars = 6
)

// mintTaskID returns a time-ordered, collision-resistant task id of the form
// "<prefix>-<time><rand>" in lowercase Crockford base32, e.g. "PROJ-01j8x2k7q7f3az".
//
// The time prefix encodes at.UnixMilli() big-endian, so ids sort lexically by creation time;
// the crypto/rand tail breaks ties within the same millisecond. Crucially, minting reads and
// writes no shared counter, so two people creating tasks in separate clones never collide on
// the id or on the on-disk filename — the conflict that a monotonic counter made unavoidable.
func mintTaskID(prefix string, at time.Time) (string, error) {
	v := uint64(0)
	if ms := at.UnixMilli(); ms > 0 {
		v = uint64(ms)
	}
	out := make([]byte, idTimeChars+idRandChars)
	for i := idTimeChars - 1; i >= 0; i-- {
		out[i] = crockford[v&31]
		v >>= 5
	}
	rb := make([]byte, idRandChars)
	if _, err := rand.Read(rb); err != nil {
		return "", fmt.Errorf("store: generate task id: %w", err)
	}
	for i, b := range rb {
		out[idTimeChars+i] = crockford[b&31] // 256 is a multiple of 32, so b&31 is unbiased
	}
	return fmt.Sprintf("%s-%s", prefix, out), nil
}
