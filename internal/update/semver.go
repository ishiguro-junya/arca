package update

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var semverPattern = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$`)

type semanticVersion struct {
	major int
	minor int
	patch int
	pre   string
}

func parseSemver(value string) (semanticVersion, error) {
	matches := semverPattern.FindStringSubmatch(strings.TrimSpace(value))
	if matches == nil {
		return semanticVersion{}, fmt.Errorf("バージョンがSemantic Versioning形式ではありません: %s", value)
	}
	parts := [3]int{}
	for i := range 3 {
		parsed, err := strconv.Atoi(matches[i+1])
		if err != nil {
			return semanticVersion{}, fmt.Errorf("バージョンを解釈できません: %w", err)
		}
		parts[i] = parsed
	}
	return semanticVersion{major: parts[0], minor: parts[1], patch: parts[2], pre: matches[4]}, nil
}

func normalizeSemver(value string) (string, error) {
	parsed, err := parseSemver(value)
	if err != nil {
		return "", err
	}
	normalized := fmt.Sprintf("%d.%d.%d", parsed.major, parsed.minor, parsed.patch)
	if parsed.pre != "" {
		normalized += "-" + parsed.pre
	}
	return normalized, nil
}

func compareSemver(left, right string) (int, error) {
	a, err := parseSemver(left)
	if err != nil {
		return 0, err
	}
	b, err := parseSemver(right)
	if err != nil {
		return 0, err
	}
	for _, pair := range [][2]int{{a.major, b.major}, {a.minor, b.minor}, {a.patch, b.patch}} {
		if pair[0] < pair[1] {
			return -1, nil
		}
		if pair[0] > pair[1] {
			return 1, nil
		}
	}
	return comparePrerelease(a.pre, b.pre), nil
}

func comparePrerelease(left, right string) int {
	if left == right {
		return 0
	}
	if left == "" {
		return 1
	}
	if right == "" {
		return -1
	}
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	for i := range max(len(leftParts), len(rightParts)) {
		if i >= len(leftParts) {
			return -1
		}
		if i >= len(rightParts) {
			return 1
		}
		if comparison := compareIdentifier(leftParts[i], rightParts[i]); comparison != 0 {
			return comparison
		}
	}
	return 0
}

func compareIdentifier(left, right string) int {
	leftNumber, leftErr := strconv.Atoi(left)
	rightNumber, rightErr := strconv.Atoi(right)
	switch {
	case leftErr == nil && rightErr == nil:
		if leftNumber < rightNumber {
			return -1
		}
		if leftNumber > rightNumber {
			return 1
		}
		return 0
	case leftErr == nil:
		return -1
	case rightErr == nil:
		return 1
	default:
		return strings.Compare(left, right)
	}
}

func acceptsPrerelease(version string) bool {
	parsed, err := parseSemver(version)
	if err != nil {
		return false
	}
	channel, _, _ := strings.Cut(parsed.pre, ".")
	return channel == "alpha" || channel == "beta" || channel == "rc"
}
