//go:generate stringer  -linecomment=false -output=generated_test.go -type=pStage,ItemFL,LintFL
// Copyright 2019 Wojciech S. Czarnecki, OHIR-RIPE. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

package octok

import (
	"bytes"
	"fmt"
)

// bullets ➊ ➋ ➌ ➍ ➎ ➏ ➐ ➑ ➒ ➓
func (v *OcItem) String() string {
	return fmt.Sprintf("--- OcLine %p ---\n\tflag: %s\n\tnSta: %2d\n\tnEnd: %2d\n\tvSta: %2d\n\tvEnd: %2d\n\tpSta: %2d\n\tmSta: %2d\n\tpEnd: %2d\n\ttype: %02d\n\tnPar: %04x\n",
		v, v.Fl.StrAll(), v.Ns, v.Ne, v.Vs, v.Ve, v.Ps, v.Ms, v.Pe, v.Tc, v.Np)
}

func (i LintFL) StrAll() (o string) {
	var got bool
	if i == 0 {
		o = "LintOK"
		return
	}
	for n := uint32(1); n != 0; n <<= 1 { // LSb first
		if uint32(i)&n != 0 {
			if got {
				o += "|" + (i & LintFL(n)).String()
			} else {
				o += (i & LintFL(n)).String()
				got = true
			}
		}
	}
	return
}

func (i ItemFL) StrAll() (o string) {
	var got bool
	//for n := byte(0x80); n != 0; n >>= 1 { // MSb first
	for n := byte(1); n != 0; n <<= 1 { // LSb first
		if byte(i)&n != 0 {
			if got {
				o += "|" + (i & ItemFL(n)).String()
			} else {
				o += (i & ItemFL(n)).String()
				got = true
			}
		}
	}
	return
}

func mkFullConf() []byte {
	r := make([]byte, 0, 1<<15)
	for _, v := range tokTestTable {
		switch v.desc[0] {
		case '!': // skip this position
			continue
		case '@': // end of table after this
			r = append(r, v.from[:]...)
			return r
		}
		r = append(r, v.from[:]...)
	}
	return r
}

type rPrintDiffMarkers struct{ Fl, Ns, Ne, Vs, Ve, Ps, Ms, Pe, Tc, Np, Li byte }
type rOut struct{ Fl, Ns, Ne, Vs, Ve, Ps, Ms, Pe, Tc, Np byte }

func rPrintCompare(tn int, desc, mock string, oc *OcFlat, m, r OcItem,
	ml LintFL, pl OcLint, msg string, prn bool) (rstr string, ok bool) {
	// m mock, r result.
	bi := oc.Inbuf
	v := rPrintDiffMarkers{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
	// cFl, cNs, cNe, cVs, cVe, cPs, cMs, cPe, cTc, cNp
	//v.Fl, v.Ns, v.Ne, v.Vs, v.Ve, v.Ps, v.Ms, v.Pe, v.Tc, v.Np
	ok = (m == r && ml == pl.What)
	if !prn {
		return
	}
	if !ok {
		if m.Fl != r.Fl {
			v.Fl = '!'
		}
		if m.Ns != r.Ns {
			v.Ns = '!'
		}
		if m.Ne != r.Ne {
			v.Ne = '!'
		}
		if m.Vs != r.Vs {
			v.Vs = '!'
		}
		if m.Ve != r.Ve {
			v.Ve = '!'
		}
		if m.Ps != r.Ps {
			v.Ps = '!'
		}
		if m.Ms != r.Ms {
			v.Ms = '!'
		}
		if m.Pe != r.Pe {
			v.Pe = '!'
		}
		if m.Tc != r.Tc {
			v.Tc = '!'
		}
		if m.Np != r.Np {
			v.Np = '!'
		}
		if ml != pl.What {
			v.Li = '!'
		}
		//v.Fl, v.Ns, v.Ne, v.Vs, v.Ve, v.Ps, v.Ms, v.Pe, v.Tc, v.Np
		// '<,'>s#\(..\), #if m.\1 != r.\1 {^M v.\1 = '!'^M}^M#g
		// Name Value Pragma Meta Flags
	}
	var pstr string
	if r.Np != 0 {
		pstr += "\n           name parts |"
		pb := r.Ns
		x := r.Np << 1
		for x > 0 {
			e := (x & (0x1f << 11)) >> 11
			if e > 1 { // sift out sentinel 1
				nx := r.Ns + uint32(e)
				pstr += string(bi[pb : nx-1])
				pstr += "⬩"
				pb = nx
			}
			x <<= 5
		}
		pstr += string(bi[pb:r.Ne])
		pstr += "|"
	}
	b := bytes.ReplaceAll(bi, []byte("\r\n"), []byte("ø"))
	b = bytes.ReplaceAll(b, []byte("\n"), []byte("$"))
	tstr := string(b)
	var truler, mruler string
	truler = ruler[:len(bi)+2]
	for _, c := range bi {
		if c > 127 {
			truler = descBLEN(b, false)
			mruler = "\n                      ›" +
				descBLEN([]byte(mock), true)
			break
		}
	}
	return fmt.Sprintf(fms, tn, desc, msg, // fms is defined at the end of file
		v.Li, uint16(pl.What), uint16(ml), pl.What.StrAll(),
		v.Li, ml.StrAll(),
		v.Fl, byte(r.Fl), byte(m.Fl), r.Fl.StrAll(),
		v.Fl, m.Fl.StrAll(),
		v.Ns, r.Ns, m.Ns, prn32(b, r.Ns, r.Ne),
		v.Ne, r.Ne, m.Ne, prn32(b, m.Ns, m.Ne),
		v.Vs, r.Vs, m.Vs, prn32(b, r.Vs, r.Ve),
		v.Ve, r.Ve, m.Ve, prn32(b, m.Vs, m.Ve),
		v.Ps, r.Ps, m.Ps, prn32(b, r.Ps, r.Ms),
		v.Pe, prn32(b, m.Ps, m.Ms),
		v.Ms, r.Ms, m.Ms, prn32(b, r.Ms, r.Pe),
		v.Pe, prn32(b, m.Ms, m.Pe),
		v.Pe, r.Pe, m.Pe, truler,
		v.Tc, r.Tc, m.Tc, tstr, pstr,
		v.Np, r.Np, m.Np, mock, mruler,
	), ok
}

func prn32(b []byte, fro, to uint32) (r []byte) {
	l := uint32(len(b))
	if l > 0 && to <= l && to > fro {
		r = b[fro:to]
	}
	return
}
func getToParse(tv ptestItem) (b []byte, ok bool, rerr string) {
	//fmt.Printf("XXX \n")
	CRLF := []byte{0xc3, 0xb8} // ø
	crlf := []byte{0xd, 0xa}
	lf := []byte{0xa}
	NL := []byte{'$'}
	TA := []byte{'!'}
	ta := []byte{0x9}
	DEL := []byte{'&'}
	del := []byte{0x7f}
	if tv.desc[0] == '^' || tv.desc[1] == '^' {
		NL[0] = '^'
	}
	bn := bytes.ReplaceAll([]byte(tv.from), NL, lf)
	b = bytes.ReplaceAll(bn, CRLF, crlf)
	b = bytes.ReplaceAll(b, TA, ta)
	b = bytes.ReplaceAll(b, DEL, del)
	ok = true
	return
}

func toHex(s string) (r string) {
	for _, c := range []byte(s) {
		r += fmt.Sprintf("%02x ", c)
	}
	return
}
func getExpected(tv ptestItem) (mre OcItem, ok bool, rerr string) {
	if len(tv.from) != len(tv.mock) {
		rerr = fmt.Sprintf("From and Mock byte lengths differ %d!=%d!\n %s\n %s\n %s\n %s\n",
			len(tv.from), len(tv.mock), descBLEN([]byte(tv.from), false), tv.from, tv.mock, descBLEN([]byte(tv.mock), true))
		return
	}
	var p uint32
	if tv.pres == ParseNone {
		ok = true // give empty mre
		return
	}
	bm := []byte(" " + tv.mock) // shift then use unfilled (zero) as flag
	lm := uint32(len(bm))
	// make mre from mock
	for ; p < lm; p++ {
		var adj uint32
		c := bm[p]
		if c > 127 { // utf8, get to the last byte
			for c <<= 1; c > 127; c <<= 1 {
				adj++
				p++
			}
			c = bm[p]
		}
		switch c {
		case 0x20, 0xb7, 0xbe, 0xbf, 0x8b: // filler
			continue
		case 'N', 0x98, 0xa6, 0xae, 0x81: // Ns
			mre.Ns = p - adj
		case 'n', 0xb8, 0xa4, 0x8e, 0x9b: // Ne
			if mre.Ns == 0 { // single char or rune
				mre.Ns = p - adj
			}
			mre.Ne = p + 1
		case 0x27: // Forced name start
			mre.Ns = p + 1
			mre.Ne = p + 1
		case ':': // separator position
			switch {
			case mre.Ns == 0: // ORD item (no name)
				mre.Ns = p
				mre.Ne = p
			}
		case 'E': // empty value position
			mre.Vs = p
			mre.Ve = p
		case 'V', 0xa8, 0xb3, 0xb6, 0x89: // Vs
			mre.Vs = p - adj
		case 'v', 0x88, 0xb2, 0x96, 0xa3: // Ve
			if mre.Vs == 0 { // single char or rune
				mre.Vs = p - adj
			}
			mre.Ve = p + 1
		case 'P': // Ps
			if mre.Ve == 0 { // empty value
				mre.Vs = p
				mre.Ve = p
			}
			mre.Ps = p
		case 'M': // Ms
			if mre.Ve == 0 { // empty value
				mre.Vs = p
				mre.Ve = p
			}
			if mre.Ps == 0 {
				mre.Ps = p
			}
			mre.Ms = p
		case '^', 0xa7: // ^§ EoI
			mre.Pe = p - adj
			if mre.Vs == 0 { // empty value
				mre.Vs = mre.Pe
				mre.Ve = mre.Pe
			}
			if mre.Ps == 0 { // no pragma
				mre.Ps = mre.Pe
			}
			if mre.Ms == 0 { // no meta
				mre.Ms = mre.Pe
			}
		default:
			// error in mock string
			rerr = fmt.Sprintf("Unknown marker %02x @%03d of mock string!", c, p)
			return
		}
	}
	if mre.Pe == 0 { // bad mock
		rerr = fmt.Sprintf("Mock string bad! No EoI marker given!")
	}
	mre.Fl = tv.fl
	mre.Tc = tv.tt
	mre.Np = tv.np
	// adjust, we shifted inputs to +1
	if mre.Ne > 0 { // panic fuse
		mre.Ns--
		mre.Ne--
		mre.Vs--
		mre.Ve--
		mre.Ps--
		mre.Ms--
		mre.Pe--
		ok = true
	} else {
		rerr = fmt.Sprintf("Mock parsed wrong!")
		mre = OcItem{}
	}
	return
}

func descBLEN(b []byte, up bool) (r string) {
	dx := [5]rune{'₁', '₂', '₃', '₄', '‗'}
	ux := [5]rune{'¹', '²', '³', '⁴', '‾'}
	dt := dx[:]
	if up {
		dt = ux[:]
	}
	max := len(b)
	for p := 0; p < max; {
		var a uint32
		var w string
		c := b[p]
		// Simplified wide print chars (Asian, e2,ba .. e9) check
		if c&0xF0 == 0xE0 && p < max-2 { // check wide print
			n, nn := b[p+1], b[p+2]
			switch {
			case c == 0xE2 && n > 0xB9,
				c > 0xE2 && c < 0xEA,                // asian sure
				c == 0xEF && n == 0xBC,              // fullwidth ！.. ＿
				c == 0xEF && n == 0xBD && nn < 0xA1: // fullwidth ｀.. <U+FF80>
				w += " "
			}
		}
		if c > 127 {
			for c <<= 1; c > 127; c <<= 1 {
				a++
				p++
			}
		}
		w += string(dt[a])
		r += w
		p++
	}
	r += " B/rune"
	return
}

/*
// func MeasureUtf takes string and measure multibyte runes in it then
// returns strings that describe byte lengths and position of each rune
// TODO - leave alone this subyak hairy, please
func MeasureUtf(s string) (cpos, rbl, oru, dru string) {
	var ms bool
	ble := "¹²³⁴"
	sup := "⁰¹²³⁴⁵⁶⁷⁸⁹"
	sub := "₀₁₂₃₄₅₆₇₈₉"
	for _, c := range []byte(s) {
		if c > 127 {
			ms = true
			break
		}
	}
	return // ruler[:32]
}
func getRuler(b, pre string) (s, rbl, oru, dru string) {
	return
}
*/
// end of fancy rulers Yak.

//const ruler = "0123456789¹123456789²123456789³123456789⁴123456789⁵123456789⁶"
const ruler = "0123456789¹123456789²123456789³123456789⁴123456789⁵123456789⁶123456789⁷123456789⁸123456789⁹123456789⁰123456789¹1234567"

const shline = "—————————————————————————————————————————————————————————————————————————\n"
const shruler = `—————————————————————————————————————————————————————————————————————————
0123456789¹123456789²123456789³123456789⁴123456789⁵123456789⁶123456789⁷12
`
const fms string = `-- TestsTable line %d --[%s]-- %s
        Parsed : Expected  Strings
  %c lint:  %03x : %03x    p: %s
  %c            :        e: %s
  %c flag:   %02x : %02x     p: %s
  %c            :        e: %s
  %c   Ns:   %02d : %02d     p: ›%s‹
  %c   Ne:   %02d : %02d ___ e: »%s«
  %c   Vs:   %02d : %02d     p: ›%s‹
  %c   Ve:   %02d : %02d ___ e: »%s«
  %c   Ps:   %02d : %02d     p: ›%s‹
  %c            :    ___ e: »%s«
  %c   Ms:   %02d : %02d     p: ›%s‹
  %c            :        e: »%s«
  %c   Pe:   %02d : %02d   ›%s
  %c   Tc:   %02x : %02x   ›%s‹%s
  %c   Np: %04x : %04x ›%s‹%s
`
