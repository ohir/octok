package octok

import (
	"strings"
	"testing"
)

const tConf string = `
%oconf hints line is first
! " / are comment lines
" below are shown common line pragmas:
$ dollar
% percent
& ampersand
* asteriks
+ plus
, comma
- hyphen
. dot
" ( ) and / can be defined as line pragmas too
" but it is discouraged
name : value
`

type LPHandPars struct {
	c, cwas, cread byte
	atpos          int
	buf            []byte
}

func LPHand(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
	if r, ok := fpar.(*LPHandPars); ok {
		r.cwas = pch
		r.atpos = oc.Inpos
		r.cread = oc.Inbuf[oc.Inpos]
		return true
	}
	return
}

const prstr string = `%$&*+,-.` // $ % & * + , - . /
func TestPragmaHandlerRegistering(t *testing.T) {
	var pRet [9]LPHandPars
	var oc OcFlat
	for n, c := range []byte(prstr) {
		if ok := RegisterLinePragma(c, &oc, LPHand, &pRet[n]); !ok {
			t.Errorf("Bad. %c should register but it did not!", c)
		}
		if ok := RegisterLinePragma(c, &oc, LPHand, &pRet[n]); ok {
			t.Errorf("Bad. '%c' should not register twice but it did!", c)
		}
	}
	if ok := RegisterLinePragma('(', &oc, LPHand, &pRet[8]); ok {
		t.Errorf("Bad. Nineth handler should not register but it did!")
	}
}

func TestFailedPragma(t *testing.T) {
	var oc OcFlat
	var bh = func(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
		if _, ok := fpar.(*LPHandPars); ok {
			if pch == '*' {
				return false
			}
			return true
		}
		return
	}
	oc.LintFull = true
	oc.Inbuf = []byte(": \n * line pragma\n")
	var pRet LPHandPars
	if ok := RegisterLinePragma('*', &oc, bh, &pRet); !ok {
		t.Errorf("Bad. '*' handler should register but it did not!")
	}
	if ok := oc.Tokenize(); ok {
		t.Errorf("Bad. Failed pragma should Err but it parsed and fired OK.!")
	}
}

func TestFailedPragmaLint(t *testing.T) {
	var oc OcFlat
	var bh = func(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
		if _, ok := fpar.(*LPHandPars); ok {
			if pch == '*' {
				return false
			}
			return true
		}
		return
	}
	oc.LintFull = true
	oc.Inbuf = []byte(": \n * line pragma\n")
	var pRet LPHandPars
	if ok := RegisterLinePragma('*', &oc, bh, &pRet); !ok {
		t.Errorf("Bad. '*' handler should register but it did not!")
	}
	if ok := TokenizeLint(&oc); ok {
		t.Errorf("Bad. Failed pragma should Err but it parsed and fired OK.!")
	}
}

func TestModifyingLinePragma(t *testing.T) {
	var oc OcFlat
	var tstr = "!! just a comment \n. include pragma\n"
	oc.Inbuf = []byte(tstr)
	var pRet LPHandPars
	pRet.buf = []byte(" name : value \n")
	if ok := RegisterLinePragma('.', &oc, includeHandler, &pRet); !ok {
		t.Errorf("Bad. '.' handler should register but it did not!")
	}
	if ok := oc.Tokenize(); !ok {
		t.Errorf("Modified buffer should parse but it did not! BadLint: %s!\nFrom: »%s«",
			oc.BadLint.What.StrAll(), oc.Inbuf)
	}
	if len(oc.Items) < 1 {
		t.Errorf("No results for OK parse from modified buffer: »%s«", oc.Inbuf)
	} else if string(oc.Inbuf[oc.Items[0].Vs:oc.Items[0].Ve]) != "value" {
		t.Errorf("Bad parse results from modified buffer: »%s«", oc.Inbuf)
		//} else {
		//	t.Errorf("Parsed OK from modified buffer: »%s«", oc.Inbuf)
	}
	oc = OcFlat{} // handlerBadBufLonger
	oc.Inbuf = []byte(tstr)
	if ok := RegisterLinePragma('.', &oc, handlerBadBufLonger, &pRet); !ok {
		t.Errorf("Bad. '.' handler should register but it did not!")
	}
	if ok := oc.Tokenize(); ok {
		t.Errorf("BufLonger should Err but it did not!")
	}
	oc = OcFlat{} // handlerBadBufShorter,
	oc.Inbuf = []byte(tstr)
	if ok := RegisterLinePragma('.', &oc, handlerBadBufShorter, &pRet); !ok {
		t.Errorf("Bad. '.' handler should register but it did not!")
	}
	if ok := oc.Tokenize(); ok {
		t.Errorf("BufShorter should Err but it did not!")
	}
}

func TestModLinePragmaLiner(t *testing.T) {
	var oc OcFlat
	var tstr = "!! just a comment \n. include pragma\n"
	oc.Inbuf = []byte(tstr)
	var pRet LPHandPars
	pRet.buf = []byte(" name : value \n")
	if ok := RegisterLinePragma('.', &oc, includeHandler, &pRet); !ok {
		t.Errorf("Bad. '.' handler should register but it did not!")
	}
	if ok := TokenizeLint(&oc); !ok {
		t.Errorf("Modified buffer should parse but it did not! BadLint: %s!\nFrom: »%s«",
			oc.BadLint.What.StrAll(), oc.Inbuf)
	}
	if len(oc.Items) < 1 {
		t.Errorf("No results for OK parse from modified buffer: »%s«", oc.Inbuf)
	} else if string(oc.Inbuf[oc.Items[0].Vs:oc.Items[0].Ve]) != "value" {
		t.Errorf("Bad parse results from modified buffer: »%s«", oc.Inbuf)
		//} else {
		//	t.Errorf("Parsed OK from modified buffer: »%s«", oc.Inbuf)
	}
	oc = OcFlat{} // handlerBadBufLonger
	oc.Inbuf = []byte(tstr)
	if ok := RegisterLinePragma('.', &oc, handlerBadBufLonger, &pRet); !ok {
		t.Errorf("Bad. '.' handler should register but it did not!")
	}
	if ok := TokenizeLint(&oc); ok {
		t.Errorf("BufLonger should Err but it did not!")
	}
	oc = OcFlat{} // handlerBadBufShorter,
	oc.Inbuf = []byte(tstr)
	if ok := RegisterLinePragma('.', &oc, handlerBadBufShorter, &pRet); !ok {
		t.Errorf("Bad. '.' handler should register but it did not!")
	}
	if ok := TokenizeLint(&oc); ok {
		t.Errorf("BufShorter should Err but it did not!")
	}
}

func TestSmallBuffer(t *testing.T) {
	var oc OcFlat
	oc = OcFlat{LintFull: true, Inbuf: []byte("")}
	if ok := oc.Tokenize(); ok {
		t.Errorf("Bad. Empty buffer should not parse OK but it did!")
	}
	oc = OcFlat{LintFull: true, Inbuf: []byte("")}
	if ok := TokenizeLint(&oc); ok {
		t.Errorf("Bad. Empty buffer should not lint OK but it did!")
	}

	oc = OcFlat{LintFull: true, Inbuf: []byte(":")}
	if ok := oc.Tokenize(); ok {
		t.Errorf("Bad. Single char buffer should not parse OK but it did!")
	}
	oc = OcFlat{LintFull: true, Inbuf: []byte(":")}
	if ok := TokenizeLint(&oc); ok {
		t.Errorf("Bad. Single char buffer should not lint OK but it did!")
	}

	oc = OcFlat{LintFull: true, Inbuf: []byte(":\n")}
	if ok := oc.Tokenize(); !ok {
		t.Errorf("Bad. Single empty ORD should parse OK but it did not!")
	}
	oc = OcFlat{LintFull: true, Inbuf: []byte(":\n")}
	if ok := TokenizeLint(&oc); !ok {
		t.Errorf("Bad. Single empty ORD should lint OK but it did not!")
	}

	oc = OcFlat{LintFull: true, Inbuf: []byte(":  ")}
	if ok := oc.Tokenize(); ok {
		t.Errorf("Bad. Buffer with no ending NL should not parse OK but it did!")
	}
	oc = OcFlat{LintFull: true, Inbuf: []byte(":  ")}
	if ok := TokenizeLint(&oc); ok {
		t.Errorf("Bad. Buffer with no ending NL should not lint OK but it did!")
	}

	oc = OcFlat{LintFull: true, Inbuf: []byte("name : value\n: val /")}
	if ok := oc.Tokenize(); ok {
		t.Errorf("Bad. No ending NL should never parse OK but it did!")
	}
	oc = OcFlat{LintFull: true, Inbuf: []byte("name : value\n: val /")}
	if ok := TokenizeLint(&oc); ok {
		t.Errorf("Bad. No ending NL should never parse nor lint OK but it did!")
	}
}

// TODO make loop for every single type / meta chars
func TestLinterRestricted(t *testing.T) {
	var oc OcFlat
	if ok := LinterSetup(&oc, LinterPragmaChars{"_`%+|'", "$#?", ">])", ">)"}); !ok {
		t.Errorf("Bad. Linter setup unexpectedly failed!")
	}
	if ok := LinterSetup(&oc, LinterPragmaChars{"_`%+|'?", "$#?", ">])", "}]"}); ok {
		t.Errorf("Bad. Linter setup succeeded while it should NOT! (? in pragmas set)")
	}
	if ok := LinterSetup(&oc, LinterPragmaChars{"_`%+|'", "$#?+", ">])", "]>"}); ok {
		t.Errorf("Bad. Linter setup succeeded while it should NOT! (+ in types set)")
	}
	if ok := LinterSetup(&oc, LinterPragmaChars{"_`%+|'", "$#?", ">])|", "])"}); ok {
		t.Errorf("Bad. Linter setup succeeded while it should NOT! (| in metas set)")
	}
	if ok := LinterSetup(&oc, LinterPragmaChars{"_`%+|'", "$#?", ">])", ";)"}); ok {
		t.Errorf("Bad. Linter setup succeeded while it should NOT! (; in Vspecials set)")
	}
}

// TODO make this to table
func TestTokenizeKnobs(t *testing.T) {
	var oc OcFlat
	oc.NoTypes = true
	oc.Inbuf = []byte("name : value '<tag>.\n")
	if ok := oc.Tokenize(); !ok || oc.LapsesFound != 0 {
		t.Errorf("Bad. '<tag>. pragma should parse with NoTypes set but it did not! [G:%d]", oc.LapsesFound)
	}
	if ok := TokenizeLint(&oc); !ok || oc.LapsesFound != 0 {
		t.Errorf("Bad. '<tag>. pragma should lint with NoTypes set but it did not! [G:%d]", oc.LapsesFound)
	}
	oc = OcFlat{}
	oc.NoTypes = true
	oc.Inbuf = []byte("name : value '?.\n")
	if ok := oc.Tokenize(); !ok || oc.LapsesFound != 1 {
		t.Errorf("Bad. '?. pragma should NOT parse with NoTypes set but it did! [G:%d]", oc.LapsesFound)
	}
	if ok := TokenizeLint(&oc); !ok || oc.LapsesFound != 1 {
		t.Errorf("Bad. '?. pragma should NOT lint with NoTypes set but it did! [G:%d]", oc.LapsesFound)
	}
	oc = OcFlat{}
	oc.NoMetas = true
	oc.Inbuf = []byte("name : value '?.\n")
	if ok := oc.Tokenize(); !ok || oc.LapsesFound != 0 {
		t.Errorf("Bad. '?. pragma should parse with NoMetas set but it did not! [G:%d]", oc.LapsesFound)
	}
	if ok := TokenizeLint(&oc); !ok || oc.LapsesFound != 0 {
		t.Errorf("Bad. '?. pragma should lint with NoMetas set but it did not! [G:%d]", oc.LapsesFound)
	}
	oc = OcFlat{}
	oc.NoMetas = true
	oc.Inbuf = []byte("name : value '<tag>.\n")
	if ok := oc.Tokenize(); !ok || oc.LapsesFound != 1 {
		t.Errorf("Bad. '<tag>. pragma should NOT parse with NoMetas set but it did! [G:%d]", oc.LapsesFound)
	}
	if ok := TokenizeLint(&oc); !ok || oc.LapsesFound != 1 {
		t.Errorf("Bad. '<tag>. pragma should NOT lint with NoMetas set but it did! [G:%d]", oc.LapsesFound)
	}
}

func TestPragmaCall(t *testing.T) {
	var oc OcFlat
	oc.Inbuf = []byte(tConf)
	var ocL OcFlat
	ocL.Inbuf = []byte(tConf)
	var pRet [10]LPHandPars
	var pRetL [10]LPHandPars
	if ok := RegisterLinePragma('^', &oc, LPHand, &pRet[8]); ok {
		t.Errorf("Bad. ^ should not register but it did!")
	}
	for n, c := range []byte(prstr) {
		pRet[n].c = c
		pRetL[n].c = c
		if ok := RegisterLinePragma(c, &oc, LPHand, &pRet[n]); !ok {
			t.Errorf("Bad. %c should register but it did not! (tok)", c)
		}
		if ok := RegisterLinePragma(c, &ocL, LPHand, &pRetL[n]); !ok {
			t.Errorf("Bad. %c should register but it did not! (lint)", c)
		}
	}
	chint := oc.linePragmas.lpchar
	chstr := []byte(prstr)
	for n, c := range chstr {
		if byte(chint&255) != c {
			t.Errorf("Bad. RegisterLinePragma messed up (tok)! [%c] != [%c] @%d", c, byte(chint&255), n)
		}
		chint >>= 8
	}
	chint = ocL.linePragmas.lpchar
	for n, c := range chstr {
		if byte(chint&255) != c {
			t.Errorf("Bad. RegisterLinePragma messed up (lint)! [%c] != [%c] @%d", c, byte(chint&255), n)
		}
		chint >>= 8
	}
	if ok := oc.Tokenize(); !ok {
		t.Errorf("Bad. Pragmas won't parse (tok)! %s", oc.BadLint.What.StrAll())
	}
	if len(oc.Items) != 1 {
		t.Errorf("Bad. Nothing parsed (tok)!")
	}
	for n, s := range pRet {
		if s.c != s.cwas {
			t.Errorf("Bad. Handler was not called for [%c] @%d (tok)!", s.c, n)
		}
	}
	for n, s := range pRet {
		if s.c != s.cread {
			t.Errorf("Bad. Handler was called in bad place. Got [%c] instead of [%c] @%d (tok)!", s.cread, s.c, n)
		}
	}
	if ok := TokenizeLint(&ocL); !ok {
		t.Errorf("Bad. Pragmas won't parse (lint)!")
	}
	if len(ocL.Items) != 1 {
		t.Errorf("Bad. Nothing parsed (lint)!")
	}
	for n, s := range pRetL {
		if s.c != s.cwas {
			t.Errorf("Bad. Handler was not called for [%c] @%d (lint)!", s.c, n)
		}
	}
	for n, s := range pRetL {
		if s.c != s.cread {
			t.Errorf("Bad. Handler was called in bad place. Got [%c] instead of [%c] @%d (lint)!", s.cread, s.c, n)
		}
	}
}

func TestTokenize(t *testing.T) {
	//func TestXXX(t *testing.T) {
	var oc OcFlat
	var bad bool
	var erta []string
	var pRet [8]LPHandPars
	for n, c := range []byte(prstr) {
		pRet[n].c = c
		if ok := RegisterLinePragma(c, &oc, LPHand, &pRet[n]); !ok {
			t.Errorf("Bad. %c should register but it did not!", c)
		}
	}
	for tno, tv := range tokTestTable {
		var oc OcFlat
		var expected, parsed OcItem
		var linted OcLint
		var rerr string
		var ok bool
		tn := (tno * 3) + ttableFirstItem // line of the zero tv of the testtable
		if tv.desc[0] == '!' {            // skip this
			t.Logf("Warning!\n!desc is set on %d '%s'. This test was skipped!", tn, tv.desc)
			continue
		}
		if len(tv.from) != len(tv.mock) {
			t.Errorf("Bad test %d '%s'", tn, tv.desc)
		}
		if expected, ok, rerr = getExpected(tv); !ok {
			t.Errorf("Bad 'mock' in %d '%s' [%s]", tn, tv.desc, rerr)
		}
		if oc.Inbuf, ok, rerr = getToParse(tv); !ok {
			t.Errorf("Bad 'from' in %d '%s' [%s]", tn, tv.desc, rerr)
		}
		oc.LintFull = true
		ok = oc.Tokenize()
		if !ok {
			rerr = oc.BadLint.What.StrAll()
			if tv.pres != ParseNone {
				t.Errorf("Test %d '%s' should parse but it did not! BadLint: %s!", tn, tv.desc, rerr)
			}
			if tv.lint != oc.BadLint.What {
				t.Errorf("Test %d '%s' !ok %s (should be %s)!", tn, tv.desc, rerr, tv.lint.StrAll())
			}
			continue
		}
		ilt := len(oc.Lapses) // lints found
		ipa := len(oc.Items)  // items parsed
		var pl LintFL
		if ilt != 0 {
			pl = oc.Lapses[ilt-1].What // last parsed lint
		}
		oc.Lapses = append(oc.Lapses, OcLint{0x8000, 0}, OcLint{0x8000, 0}) // no panic
		switch {
		case !ok && ilt == 0: // really bad
			t.Errorf("Bad Tokenize in %d '%s'", tn, tv.desc)
			continue
		case !ok && pl != tv.lint:
			t.Logf("Bad Last line in %d '%s' [%s m:p %s]",
				tn, tv.desc, tv.lint.StrAll(), pl.StrAll())
			tv.fl |= Modified
		}
		switch {
		case ipa != int(tv.pres): // lines parsed do not match
			bad = true
			t.Logf("Items count (%d) differs from expected (%d) at %d '%s'", ipa, tv.pres, tn, tv.desc)
			for _, x := range oc.Lapses {
				if x.Line == 0x8000 {
					break
				}
				t.Logf("Linted: %s (of %d)", x.What.StrAll(), oc.LapsesFound)
			}
		default: // LGTM
			switch ipa {
			case 1:
				parsed = oc.Items[0]
				linted = oc.Lapses[0]
			case 2, 3:
				parsed = oc.Items[1]
				linted = oc.Lapses[1]
			case 0: // empty
				break
			default: //
				bad = true
				t.Errorf("TOO MUCH PARSED (%d) at %d '%s' [expected: %d] G:%d", ipa, tn, tv.desc, tv.pres, oc.LapsesFound)
				continue
			}
			if rstr, ok := rPrintCompare(tn, tv.desc, tv.mock, &oc, expected, parsed, tv.lint, linted, "Tokenize", true); !ok {
				erta = append(erta, rstr)
				bad = true
			}
		}
		if tv.desc[0] == '@' { // this was a last test
			t.Logf("Warning!\n@desc is set on %d '%s'. ALL further tests were skipped!", tn, tv.desc)
			break
		}
	}
	if bad {
		for _, s := range erta {
			t.Logf("%s", s)
		}
		t.Errorf("\n")
	}
}

func TestTokenizeLint(t *testing.T) {
	//func TestLintXXX(t *testing.T) {
	var oc OcFlat
	var bad bool
	var erta []string
	var pRet [8]LPHandPars
	for n, c := range []byte(prstr) {
		pRet[n].c = c
		if ok := RegisterLinePragma(c, &oc, LPHand, &pRet[n]); !ok {
			t.Errorf("Bad. %c should register but it did not!", c)
		}
	}
	for tno, tv := range tokTestTable {
		var oc OcFlat
		var expected, parsed OcItem
		var linted OcLint
		var rerr string
		var ok bool
		tn := (tno * 3) + ttableFirstItem // line of the zero tv of the testtable
		if tv.desc[0] == '!' {            // skip this
			t.Logf("Warning!\n!desc is set on %d '%s'. This test was skipped!", tn, tv.desc)
			continue
		}
		if len(tv.from) != len(tv.mock) {
			t.Errorf("Bad test %d '%s'", tn, tv.desc)
		}
		if expected, ok, rerr = getExpected(tv); !ok {
			t.Errorf("Bad 'mock' in %d '%s' [%s]", tn, tv.desc, rerr)
		}
		if oc.Inbuf, ok, rerr = getToParse(tv); !ok {
			t.Errorf("Bad 'from' in %d '%s' [%s]", tn, tv.desc, rerr)
		}
		oc.LintFull = true
		ok = TokenizeLint(&oc)
		if !ok {
			rerr = oc.BadLint.What.StrAll()
			if tv.pres != ParseNone {
				t.Errorf("Test %d '%s' should parse but it did not! BadLint: %s!", tn, tv.desc, rerr)
			}
			if tv.lint != oc.BadLint.What {
				t.Errorf("Test %d '%s' !ok %s (should be %s)!", tn, tv.desc, rerr, tv.lint.StrAll())
			}
			continue
		}
		ilt := len(oc.Lapses) // lints found
		ipa := len(oc.Items)  // items parsed
		var pl LintFL
		if ilt != 0 {
			pl = oc.Lapses[ilt-1].What // last parsed lint
		}
		oc.Lapses = append(oc.Lapses, OcLint{0x8000, 0}, OcLint{0x8000, 0}) // no panic
		switch {
		case !ok && ilt == 0: // really bad
			t.Errorf("Bad Tokenize in %d '%s'", tn, tv.desc)
			continue
		case !ok && pl != tv.lint:
			t.Logf("Bad Last line in %d '%s' [%s m:p %s]",
				tn, tv.desc, tv.lint.StrAll(), pl.StrAll())
			tv.fl |= Modified
		}
		switch {
		case ipa != int(tv.pres): // lines parsed do not match
			bad = true
			t.Logf("Items count (%d) differs from expected (%d) at %d '%s'", ipa, tv.pres, tn, tv.desc)
			for _, x := range oc.Lapses {
				if x.Line == 0x8000 {
					break
				}
				t.Logf("Linted: %s (of %d)", x.What.StrAll(), oc.LapsesFound)
			}
		default: // LGTM
			switch ipa {
			case 1:
				parsed = oc.Items[0]
				linted = oc.Lapses[0]
			case 2, 3:
				parsed = oc.Items[1]
				linted = oc.Lapses[1]
			case 0: // empty
				break
			default: //
				bad = true
				t.Errorf("TOO MUCH PARSED (%d) at %d '%s' [expected: %d] G:%d", ipa, tn, tv.desc, tv.pres, oc.LapsesFound)
			}
			if rstr, ok := rPrintCompare(tn, tv.desc, tv.mock, &oc, expected, parsed, tv.lint, linted, "TokLint", true); !ok {
				erta = append(erta, rstr)
				bad = true
			}
		}
		if tv.desc[0] == '@' { // this was a last test
			t.Logf("Warning!\n@desc is set on %d '%s'. ALL further tests were skipped!", tn, tv.desc)
			break
		}
	}
	if bad {
		for _, s := range erta {
			t.Logf("%s", s)
		}
		t.Errorf("\n")
	}
}

// test if base Tokenize and linting/simpified versions are in sync.
func TestRangeChecks(t *testing.T) {
	var oc OcFlat
	oc.Inbuf = []byte(" n : v \n")
	_ = TokenizeLint(&oc) // init linter checkers

	for c := byte(1); c != 0; c++ {
		tok, lint := isPragmaChar(c), lintIsPragmaChar(c, true, &oc)
		if tok != lint {
			t.Errorf("[%02x] isPragmaChar Tok:%v != Lint(%v)\n", c, tok, lint)
		}
	}
	for c := byte(1); c != 0; c++ {
		tok, lint := isPragmaNotMeta(c), lintIsPragmaChar(c, false, &oc)
		if tok != lint {
			t.Errorf("[%02x] isPragmaNotMeta Tok:%v != Lint(%v)\n", c, tok, lint)
		}
	}
	for c := byte(1); c != 0; c++ {
		tok, lint := isValueSpecial(c), isValueSpecLint(c, &oc)
		if tok != lint {
			t.Errorf("[%02x] isValueSpecial Tok:%v != Lint(%v)\n", c, tok, lint)
		}
	}
	// isValueSpecLint
}

func TestLinterMessages(t *testing.T) {
	j := 0
	i := LintOK
	k := (LintUnknown << 1)
	if s := k.String(); !strings.HasPrefix(s, "LintFL(") {
		t.Errorf("LintUnknown is not the last flag (%d %s is?). Other tests might be wrong!", k, k.String())
	}
	if s, ok := LintMessage(k); ok {
		t.Errorf("Messed up linter messages table. Should NOT have an entry for %s! Has: %s", k.String(), s)
	}
	if s, ok := LintMessage(i); !ok || len(s) < 10 {
		t.Errorf("Bad (too short) messages table entry for %s! [%s]", i.String(), s)
	}
	for i := LintOK + 1; i <= LintUnknown; i <<= 1 {
		s, ok := LintMessage(i)
		if !ok {
			t.Errorf("Messed up linter messages table. No entry for %s found!", i.String())
		}
		if len(s) < 10 {
			t.Errorf("Too short message for %s ! [%s]", i.String(), s)
		}
		t.Logf("@%02d %s => %s", j, i.String(), s) // for -v to see
		j++
	}
}

var handlerBadPos = func(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
	oc.Inpos += 2
	return true
}
var handlerBadBufLonger = func(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
	oc.Inbuf = append(oc.Inbuf, 'x')
	return true
}
var handlerBadBufShorter = func(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
	oc.Inbuf = oc.Inbuf[:oc.Inpos]
	return true
}
var includeHandler = func(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
	if r, ok := fpar.(*LPHandPars); ok {
		if pch != '.' {
			return false
		}
		p := oc.Inpos
		n := p
		c := oc.Inbuf[n]
		for ; n < len(oc.Inbuf); n++ {
			c = oc.Inbuf[n]
			if c == '\n' {
				break
			}
		}
		if c != '\n' {
			return false
		}
		oc.Inpos = n
		var parts = [...][]byte{
			oc.Inbuf[:n+1],
			[]byte("!+++ inserted by .\n"),
			r.buf,
			[]byte("\n!--- end of . inserted\n"),
			oc.Inbuf[n:],
		}
		var cs int
		for _, k := range parts[:] {
			cs += len(k)
		}
		nbuf := make([]byte, cs, cs)
		cs = 0
		for _, k := range parts[:] {
			copy(nbuf[cs:], k)
			cs += len(k)
		}
		oc.Inbuf = nbuf
		return true
	}
	return
}

// parsed line flags
type paResult byte

// stringer: no values defined for type paResult
const ( // Parser Exits. 0-9 reserved for tests
	ParseNone paResult = 0
	ParseOK   paResult = 1 // tests: single item parsed, check [0]
	Parsed2   paResult = 2 // tests: two   items parsed, check [1]
	Parsed3   paResult = 3 // tests: three items parsed, check [1]
	ParseInitErr
)

// import fmt
// func prOcLine(l OcItem, oc *Flat) {
//	fmt.Printf("GOT: %s\n    »%s«\nNs:%02d, Ne:%02d, Vs:%02d, Ve:%02d, Ps:%02d, Ms:%02d, Pe:%02d, Np:%04x, Tc:%02x, Fl:%02x\n",
//		ruler[:74], oc.Inbuf[:len(oc.Inbuf)-1], l.Ns, l.Ne, l.Vs, l.Ve, l.Ps, l.Ms, l.Pe, l.Np, l.Tc, byte(l.Fl))
//}
