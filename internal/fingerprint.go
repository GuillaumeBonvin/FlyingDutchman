package internal

import (
	"bufio"
	"github.com/pion/sdp/v2"
	"github.com/pion/webrtc/v3"
	"math/big"
	"os"
	"strings"
)

// returns the fingerprint without the encryption method as a string from an sdp.SessionDescription
func ExtractFingerprint(desc *sdp.SessionDescription) string {
	fingerprints := []string{}

	if fingerprint, haveFingerprint := desc.Attribute("fingerprint"); haveFingerprint {
		fingerprints = append(fingerprints, fingerprint)
	}

	for _, m := range desc.MediaDescriptions {
		if fingerprint, haveFingerprint := m.Attribute("fingerprint"); haveFingerprint {
			fingerprints = append(fingerprints, fingerprint)
		}
	}

	if len(fingerprints) < 1 {
		return ""
	}

	for _, m := range fingerprints {
		if m != fingerprints[0] {
			return ""
		}
	}

	parts := strings.Split(fingerprints[0], " ")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// returns the fingerprint without the encryption method as a string from an webrtc.DTLSFingerprint
func FingerprintToString(fingerprint webrtc.DTLSFingerprint) string {
	result := strings.ToUpper(fingerprint.Value)
	return result
}

// returns a passphrase obtained by a base change of the given fingerprint value
func FingerprintToPhrase(fingerprint string) string {
	// get the dictionary as an array of words from a text file
	dictionary, err := readLines("ressources/big_dictionary.txt")
	if err != nil {
		panic(err)
	}
	// format fingerprint to hexadecimal string
	hexa := strings.ReplaceAll(fingerprint, ":", "")

	// value of fingerprint
	left := new(big.Int)
	left.SetString(hexa, 16)

	// length of dictionary
	base := new(big.Int)
	base.SetInt64(int64(len(dictionary)))

	passphrase := ""

	// for five runs, picks the word at index left%base and adds it to the passphrase
	for i := 0; i < 5; i++ {
		wordIndex := new(big.Int)
		// wordindex = left % base
		wordIndex.Rem(left, base)
		word := dictionary[wordIndex.Int64()]

		if i != 4 {
			passphrase += word + "-"
		} else {
			passphrase += word
		}

		// divides the fingerprint value by the dictionary length
		left.Div(left, base)
	}

	return passphrase
}

// generates an array of strings made from each newline of a txt file
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
