package lib

import (
	"strings"
)

/**
 * 加密解密类.
 */

type Crypto struct{}

const ralphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890_"
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890_"
const password = "kinge383e"

/**
* @param string  strtoencrypt
* @return string
 */
func (Crypto) Encrypt(strtoencrypt string) string {

	count := len(password)
	length := len(ralphabet)
	pos_alpha_ary := make([]string, count)
	for i := 0; i < count; i++ {
		index := strings.Index(alphabet, string(password[i]))
		pos_alpha_ary[i] = alphabet[index:][:length]
	}

	i, n := 0, 0
	c := len(strtoencrypt)
	str := ""

	for i < c {
		index := strings.Index(alphabet, string(strtoencrypt[i]))
		str += string(string(pos_alpha_ary[n])[index])
		n++
		if n == count {
			n = 0
		}

		i++
	}

	return str
}

/**
* @param string  strtoencrypt
* @return string
 */
func (Crypto) Decrypt(strtodecrypt string) string {
	count := len(password)
	length := len(ralphabet)

	pos_alpha_ary := make([]string, count)
	for i := 0; i < count; i++ {
		index := strings.Index(alphabet, string(password[i]))
		pos_alpha_ary[i] = alphabet[index:][:length]
	}

	i, n := 0, 0
	c := len(strtodecrypt)
	str := ""

	for i < c {
		index := strings.Index(pos_alpha_ary[n], string(strtodecrypt[i]))
		str += string(string(ralphabet[index]))

		n++

		if n == count {
			n = 0
		}

		i++
	}

	return str
}
