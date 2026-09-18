package provider

import "strings"

var romajiTable = func() map[rune]string {
	m := map[rune]string{
		'あ': "a", 'い': "i", 'う': "u", 'え': "e", 'お': "o",
		'か': "ka", 'き': "ki", 'く': "ku", 'け': "ke", 'こ': "ko",
		'さ': "sa", 'し': "shi", 'す': "su", 'せ': "se", 'そ': "so",
		'た': "ta", 'ち': "chi", 'つ': "tsu", 'て': "te", 'と': "to",
		'な': "na", 'に': "ni", 'ぬ': "nu", 'ね': "ne", 'の': "no",
		'は': "ha", 'ひ': "hi", 'ふ': "fu", 'へ': "he", 'ほ': "ho",
		'ま': "ma", 'み': "mi", 'む': "mu", 'め': "me", 'も': "mo",
		'や': "ya", 'ゆ': "yu", 'よ': "yo",
		'ら': "ra", 'り': "ri", 'る': "ru", 'れ': "re", 'ろ': "ro",
		'わ': "wa", 'を': "wo", 'ん': "n",
		'が': "ga", 'ぎ': "gi", 'ぐ': "gu", 'げ': "ge", 'ご': "go",
		'ざ': "za", 'じ': "ji", 'ず': "zu", 'ぜ': "ze", 'ぞ': "zo",
		'だ': "da", 'ぢ': "ji", 'づ': "zu", 'で': "de", 'ど': "do",
		'ば': "ba", 'び': "bi", 'ぶ': "bu", 'べ': "be", 'ぼ': "bo",
		'ぱ': "pa", 'ぴ': "pi", 'ぷ': "pu", 'ぺ': "pe", 'ぽ': "po",
		'ぁ': "a", 'ぃ': "i", 'ぅ': "u", 'ぇ': "e", 'ぉ': "o",
		'ゃ': "ya", 'ゅ': "yu", 'ょ': "yo",
		'ゔ': "vu",
	}
	out := map[rune]string{}
	for k, v := range m {
		out[k] = v
		// 平假名 -> 片假名（+0x60）
		if k >= 'ぁ' && k <= 'ゖ' {
			out[k+0x60] = v
		}
	}
	out['ー'] = "-"
	out['ヷ'] = "va"
	out['ヸ'] = "vi"
	out['ヹ'] = "ve"
	out['ヺ'] = "vo"
	return out
}()

func romajiSingle(r rune) string {
	return romajiTable[r]
}

// smallYVowel 拗音中的小写 ゃ/ゅ/ょ。
func smallYVowel(r rune) (string, bool) {
	switch r {
	case 'ゃ', 'ャ':
		return "a", true
	case 'ゅ', 'ュ':
		return "u", true
	case 'ょ', 'ョ':
		return "o", true
	}
	return "", false
}

func yoonConsonant(base string) (string, bool) {
	switch base {
	case "ki":
		return "ky", true
	case "gi":
		return "gy", true
	case "shi":
		return "sh", true
	case "chi":
		return "ch", true
	case "ji":
		return "j", true
	case "ni":
		return "ny", true
	case "hi":
		return "hy", true
	case "bi":
		return "by", true
	case "pi":
		return "py", true
	case "mi":
		return "my", true
	case "ri":
		return "ry", true
	}
	return "", false
}

// RomajiJA 假名转罗马音（Hepburn 简化版）。
func RomajiJA(kana string) string {
	var b strings.Builder
	runes := []rune(kana)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// 促音：っ/ッ
		if r == 'っ' || r == 'ッ' {
			if i+1 < len(runes) {
				next := romajiSingle(runes[i+1])
				if next != "" {
					if strings.HasPrefix(next, "ch") {
						b.WriteString("t")
					} else {
						b.WriteString(next[:1])
					}
				}
			}
			continue
		}

		// 拗音：きゃ/しゃ/ちゃ…
		if i+1 < len(runes) {
			if yv, ok := smallYVowel(runes[i+1]); ok {
				if c, ok2 := yoonConsonant(romajiSingle(r)); ok2 {
					b.WriteString(c + yv)
					i++
					continue
				}
			}
		}

		if s := romajiSingle(r); s != "" {
			b.WriteString(s)
		} else {
			b.WriteString(string(r))
		}
	}
	return b.String()
}
