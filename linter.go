// Copyright 2019 Wojciech S. Czarnecki, OHIR-RIPE. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

package octok

// func TokenizeLint is a reference tokenizer and linter. Unlike Tokenize
// method, its linting is fully customizable, ie. it allows to restrict
// set of value pragmas to exact subset used by a particular implementation;
// possibly by an implementation in a language other than Go.
func TokenizeLint(oc *OcFlat) (ok bool) {
	var nowStage, fromStage pStage // parse stages
	var afterS, lastP int          // position markers
	var culint LintFL              // current line ambigs
	var ln uint32 = 1              // current line №
	var items []OcItem             // items found
	var lapses []OcLint            // ambigs found
	var b []byte = oc.Inbuf[:]     // buffer to parse
	var p int                      // position in buffer
	var c byte                     // current char at p
	var l OcItem                   // current parses
	var gotSep, gotItem bool       // separator seen, new Item
	var gotQuote, gotCom bool      // ordinary key, Comment
	noTypes := oc.NoTypes          // wholesale knobs
	withMet := !oc.NoMetas         //
	linC := oc.linePragmas.lpchar  // line pragmas table

	blen := len(b)                 // buflen is used more than once
	if blen < 2 || blen > u32max { // nothing to parse, or too much
		oc.LapsesFound++
		oc.Lapses = append(oc.Lapses, OcLint{0, LintBadBufLen})
		return
	}
	items = make([]OcItem, 0, oc.ItemsExpected)
	lapses = make([]OcLint, 0, oc.ItemsExpected/8)
	if oc.Pck == 0 { // not configured, make full range
		oc.Pck = pragmaChars
		oc.Tck = typeChars
		oc.Mck = metaChars
		oc.Sck = specKeChars
	}

	for ; p < blen; p++ {
		c = b[p]
		switch { // loop tight on uninteresting bytes
		case c == 0x20 || c == 0x0d:
			continue
		case !gotItem:
			break
		case c == 0x0a:
			nowStage = registerItem
		case c < 0x20 || c == 0x7f:
			nowStage = badChar
			break
		case gotCom:
			continue
		case p-afterS == 1 && c != 0x2e:
			afterS++
			continue // any after space excluding dot
		case !gotSep && c == 0x3a:
			nowStage = ckSEP
		case gotSep && c&^1 != 0x2e:
			afterS++
			continue // neither dot nor slash
		}
		if c > 0x20 { // name or pragma endpos
			lastP = afterS
		}
		afterS = p

		switch nowStage {
		case inValue:
			switch {
			case blen-p < 2:
				break
			case c == '/' && b[p+1] == '/' && b[p-1] == ' ':
				l.Ve = uint32(p) // Keep at slash position.
			case b[p+1] < 0x21 && lintIsPragmaChar(b[p-1], true, oc):
				l.Pe = uint32(p)
				if l.Ve != 0 {
					culint |= LintRemCancel
				}
				l.Ve = 0 // no pragmas in remark allowed
			}
		case inName: // split name on dots and spaces
			if !gotQuote && isStructureLint(c, oc) {
				l.Fl |= IsSpec
				lastP = p
				for c == b[p] {
					p++
				}
				p--
				afterS = p
				c = '^'
			}
			if l.Np&NpOverParts != 0 || p-int(l.Ns) > 31 {
				culint |= LintKeyParts
				continue // more than 3 or part starts at offset > 31
			}
			if l.Ns == uint32(p) || int(l.Ns)-lastP == 1 { // dot or space lead
				continue
			}
			if l.Np == 0 { // set sentinel 1. Part can't have offset of 1.
				l.Np++
			}
			l.Np <<= 5
			switch c {
			case '.':
				l.Np |= uint16(p - int(l.Ns) + 1)
			case '^':
				l.Np |= uint16(lastP - int(l.Ns))
			default:
				l.Np |= uint16(p - int(l.Ns))
			}
		case lpCheck: // first non space in a line gets here
			switch {
			case c == ':':
				l.Ns = uint32(p)
				fromStage = inName // either start of a name or ORD separator
				gotItem = true
				break // fall to ckSEP
			case c == 0x27: // ' forces name start
				l.Ns = uint32(p + 1)
				nowStage = inName
				gotItem = true
				gotQuote = true
				continue
			case isStructureLint(c, oc):
				l.Fl |= IsSpec
				fallthrough
			case c > 0x2f: // Got to name's first
				if c < 0x3a && c > 0x2f { // ascii digit
					l.Fl |= IsOrd
				}
				l.Ns = uint32(p)
				nowStage = inName
				gotItem = true
				continue
			case c == '\n': // skip empty lines
				continue
			default: // line comment or line pragma.
				if c > 0x23 && linC != 0 {
					pch := linC
					for n := 0; pch > 0; pch >>= 8 {
						if c == byte(pch&255) {
							oc.Inpos = p      // let handler know position
							oc.InLine = ln    // including a line no
							blenwas := len(b) // we'll check it after
							b = nil           // release backing array
							if ok := oc.linePragmas.lpcall[n](c, oc, oc.linePragmas.lpfpar[n]); !ok {
								// if handler messed up, we can not reliably proceed
								oc.BadLint = OcLint{ln, LintBadLnPrag}
								return false
							}
							b = oc.Inbuf[:]
							blen = len(b)
							switch {
							case blen <= p || blen >= u32max:
								fallthrough
							default:
								oc.BadLint = OcLint{ln, LintBufCorrupt}
								return false
							case oc.Inpos > p && b[oc.Inpos] == '\n': // modified OK
								p = oc.Inpos - 1
								continue
							case blen == blenwas && oc.Inpos == p && b[oc.Inpos] == c:
								break
							}
							break
						} // got pragma to call
						n++
					}
				}
				gotItem = true
				gotCom = true
				continue
			}
			fallthrough
		case ckSEP:
			c = b[p+1]
			switch {
			case isStructure(c):
				l.Fl |= IsSpec
				l.Vs = uint32(p + 1)
				break
			case c < 0x20, blen-p < 4: // got empty value
				l.Vs = uint32(p + 1) // blen-p: 3210  43210  543210
				l.Ve = uint32(p + 1) // buffer:  :⬩$   : ⬩$   : .⬩$
				break
			case c != 0x20 && c != ':':
				fallthrough
			default:
				nowStage = fromStage
				continue // not a separator
			case c == 0x20,
				c == ':' && b[p+2] == ' ':
				l.Vs = uint32(p + 2)
				break
			}
			if nowStage == lpCheck { // ORD item
				l.Ne = uint32(p)
				l.Fl |= IsOrd
			} else { // NAV item
				l.Ne = uint32(lastP + 1)
			}
			gotSep = true
			nowStage = inValue
		case badChar:
			l = OcItem{}
			gotItem = true
			gotCom = true
			oc.LapsesFound++
			lapses = append(lapses, OcLint{ln, culint | LintCtlChars}) // store
			culint = 0
			continue
		case registerItem:
			nowStage = lpCheck
			gotItem = false
			if !gotSep {
				if !gotCom { // lint free comments
					culint |= LintNoComment
				}
				if culint != 0 {
					lapses = append(lapses, OcLint{ln, culint}) // store
				}
				gotCom = false
				culint = 0
				ln++
				continue
			}
			gotCom = false
			gotSep = false

			// Check for pragma, Finalize, then Register
			if l.Ne == l.Ns { // adjust from ' forced name
				l.Fl |= IsOrd
			}
			var i uint32
			var disa, guard bool
			if l.Ve > 0 { // Ve is set at first / of // remark
				i = l.Ve - 1
			} else {
				i = uint32(p - 1)
			}
			for ; i >= l.Vs; i-- { // get rid of ending space.
				c = b[i]
				if c > 0x20 {
					break
				}
			}
			if i != 0 && i == l.Pe { // Looks like a pragma, check it
				l.Pe++ // Pe needs to be right after the dot
				i--
				if withMet {
					if r, ok := lintMetaCheck(b, l.Vs, i, oc); ok {
						l.Ms = r
						i = r - 1
						if l.Ps == 0 {
							l.Ps = r
						}
						c = b[i] // could be lone meta
					}
				}
			pragmaBack:
				for ; i >= l.Vs; i-- {
					c = b[i]
					if c != ' ' && !lintIsPragmaChar(c, false, oc) { // no meta here
						break
					}
					switch c {
					case ' ': // space is the only valid start of a pragma chain
						l.Ps = i + 1
						break pragmaBack
					case '_': // filler
						for ; i > l.Vs && b[i-1] == '_'; i-- {
						}
					case '|': // guard
						guard = true
						l.Ve = i
						fallthrough
					case 0x27: // disa
						l.Ps = i
						i-- // make to possible space
						c = b[i]
						disa = true
						break pragmaBack
					case '+': // join
						l.Fl |= NextCont
						if l.Fl&NextMeta != 0 {
							culint |= LintTwoJoins
						}
					case '^': // nline
						if l.Tc&128 != 0 {
							culint |= LintDublCaret
						} else if l.Tc != 0 {
							culint |= LintTypeAndNL
							l.Tc = 0
						}
						for l.Tc = 1; i > l.Vs && b[i-1] == '^' && l.Tc < 64; i-- {
							l.Tc++
						}
						if l.Tc&TcHasErrBit != 0 {
							l.Tc = TcTooManyNL
							culint |= LintTooManyNL
							break pragmaBack
						}
						l.Tc |= TcHasCarets
					case '`': // subs
						l.Fl |= Backtick
					case 0x5c: // unesc
						l.Fl |= Unescape
					case '%': // meta join
						l.Fl |= NextMeta
						if l.Fl&NextCont != 0 {
							culint |= LintTwoJoins
						}
					default: // types fall here
						switch {
						case noTypes:
							break
						case l.Tc == 0:
							l.Tc = c
							continue
						case l.Tc&128 != 0:
							culint |= LintTypeAndNL
							l.Tc = TcTypeAndNL
						default:
							culint |= LintManyTypes
							l.Tc = TcDoublType
						}
						break pragmaBack
					}
				} // pragmaBack loop

				if c != ' ' { // pragma (chain) must start with a space.
					if i < l.Vs && lintIsPragmaChar(c, true, oc) { // even lone pragma
						l.Ps = l.Vs
						l.Ve = l.Vs
					} else { // not a pragma
						culint |= LintSusPragma
						disa = false
						l.Ve = l.Pe
						l.Ps = l.Pe
						l.Ms = l.Pe
					}
				} else if !guard { // skip spaces to the Ve.
					for ; i >= l.Vs; i-- {
						c = b[i]
						if c > 0x20 {
							break
						}
					}
					l.Ve = i + 1
				}
				if l.Ms == 0 {
					l.Ms = l.Pe
				}
			} else { // no pragma dot
				l.Pe = i + 1
				l.Ve = l.Pe
				l.Ps = l.Pe
				l.Ms = l.Pe
			}
			c = b[l.Vs]
			if l.Vs == l.Ve {
				l.Fl |= IsEmpty
			}
			if disa {
				culint &^= LintRemCancel
			}
			if culint != 0 { // store linted
				lapses = append(lapses, OcLint{ln, culint})
				culint = 0
			}
			items = append(items, l) // store item
			l = OcItem{}
			ln++
		}
	}
	oc.Items = items
	oc.Lapses = lapses
	if gotItem && !gotCom { // someone forgot to press RETURN
		oc.LapsesFound++
		oc.BadLint = OcLint{ln, LintBadEndLin}
		return false
	}
	return true
} // func TokenizeLint(oc *Parser) (ok bool)

func isStructureLint(c byte, oc *OcFlat) bool {
	if c == '^' { // section always
		return true
	}
	i := oc.Sck
	for ; i > 0; i >>= 8 {
		if c == byte(i) {
			break
		}
	}
	return i != 0
}

// func lintIsPragmaChar checks if c is a pragma character.
// This is used with reference linter
func lintIsPragmaChar(c byte, withmeta bool, oc *OcFlat) bool {
	i := oc.Pck
	for ; i > 0; i >>= 8 {
		if c == byte(i) {
			break
		}
	}
	if i == 0 {
		for i = oc.Tck; i > 0; i >>= 8 {
			if c == byte(i) {
				break
			}
		}
	}
	if i == 0 && withmeta {
		for i = oc.Mck; i > 0; i >>= 8 {
			if c == byte(i) {
				break
			}
		}
	}
	return i != 0
}

// Funtion LinterSetup prepares linter to recognize as valid only pragma
// character sets given in LinterPragmaChars struct. It does not allow
// for free changing the sets, but allows for restricting them. If any of
// provided strings contains a character that is not in the FULL set for a
// category LinterSetup will return false.
func LinterSetup(oc *OcFlat, cs LinterPragmaChars) (ok bool) {
	// configure reference linter/tokenizer
	for _, c := range []byte(cs.P) { // Pragmas string
		i := pragmaChars
		for ; i > 0; i >>= 8 {
			if c == byte(i) {
				break
			}
		}
		if i == 0 {
			return
		}
		oc.Pck <<= 8
		oc.Pck |= uint64(c)
	}
	for _, c := range []byte(cs.T) { // Types string
		i := typeChars
		for ; i > 0; i >>= 8 {
			if c == byte(i) {
				break
			}
		}
		if i == 0 {
			return
		}
		oc.Tck <<= 8
		oc.Tck |= uint64(c)
	}
	for _, c := range []byte(cs.M) { // Meta string. ENDING } ) ] > / :
		i := metaChars
		for ; i > 0; i >>= 8 {
			if c == byte(i) {
				break
			}
		}
		if i == 0 {
			return
		}
		oc.Mck <<= 8
		oc.Mck |= uint64(c)
	}
	for _, c := range []byte(cs.S) { // Spec Val string. ENDING } ) ] >
		i := specSetupCk
		for ; i > 0; i >>= 8 {
			if c == byte(i) {
				break
			}
		}
		if i == 0 {
			return
		}
		oc.Sck <<= 8
		oc.Sck |= uint64(c)
		if c == 0x29 { // )
			c++
		}
		oc.Sck <<= 8
		oc.Sck |= uint64(c - 2)
	}
	return true
}

func lintMetaCheck(b []byte, stop, i uint32, oc *OcFlat) (r uint32, ok bool) {
	var c, d, e, o byte
	r = i
again:
	o = b[i]
	x := oc.Mck
	for ; x > 0; x >>= 8 {
		if o == byte(x) {
			break
		}
	}
	if x == 0 {
		return
	}
	e = 0
	switch o {
	case ';':
		d = '@'
	case ')':
		d = '('
	case '/':
		d, e = '=', '&'
	case '>', ']', '}':
		d = o - 2
		// default: // linter version outs on x == 0 "not found"
		//	return
	}
	for i > stop { // l.Vs
		i--
		c = b[i]
		if c == d || c == e {
			ok = true
			r = i
			i--
			goto again // support multi metas
		}
	}
	return
}
