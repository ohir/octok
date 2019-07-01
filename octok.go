// Copyright 2019 Wojciech S. Czarnecki, OHIR-RIPE. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

// Octok is a tokenizer for data in the OCONF (object config) format.
// Package Octok exposes minimal API, implements all MUST and SHOULD
// elements of the OCONF line format specification so it is not expected
// to change further, except for possible bugfixes.
//
// Octok is a core of GoConf "config" package but it is published separate
// to allow easy reuse, like for serialization and for linting
// implementations done in other languages.
package octok

// Method Tokenize parses the Inbuf supposed to be in Oconf line format.
// It returns OK unless there were too few bytes to parse, last not empty
// line lacked an ending newline, or registered line pragma failed. Parsed
// Oconf ITEMs are filled into OcFlat.Items field. If (OcFlat.LintFull ==
// true), the OcFlat.Lapses table is filled. Otherwise, spotted lapses are
// simply counted in the OcFlat.LapsesFound counter.
func (oc *OcFlat) Tokenize() (ok bool) {
	var nowStage, fromStage pStage   // parse stages
	var afterS, lastP int            // position markers
	var culint LintFL                // current line lint flags
	var ln uint32 = 1                // current line №
	var items []OcItem               // items found
	var lapses []OcLint              // ambigous found
	var b []byte = oc.Inbuf[:]       // buffer to parse
	var p int                        // position in buffer
	var c byte                       // current char at p
	var l OcItem                     // current parses
	var rawB uint64                  // raw boundary
	var gotSep, gotItem, gotRaw bool // separator seen, new Item, raw
	var gotQuote, gotCom bool        // ordinary key, Comment
	noTypes := oc.NoTypes            // wholesale knobs
	rawThre := oc.RawThreshold       // 0 allows for binary raw, 255 off
	withMet := !oc.NoMetas           // localize
	lint := oc.LintFull              // fill lapses table
	linC := oc.linePragmas.lpchar    // line pragmas table

	blen := len(b)                 // buflen is used more than once
	if blen < 2 || blen > u32max { // nothing to parse, or too much
		oc.LapsesFound++
		if lint {
			oc.Lapses = append(oc.Lapses, OcLint{0, LintBadBufLen})
		}
		return
	}
	items = make([]OcItem, 0, oc.ItemsExpected)
	if lint {
		lapses = make([]OcLint, 0, oc.ItemsExpected/8)
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
		case gotCom:
			continue
		case c < 0x20 || c == 0x7f:
			nowStage = badChar
			break
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
			case b[p+1] < 0x21 && isPragmaChar(b[p-1]):
				l.Pe = uint32(p)
				if l.Ve != 0 {
					culint |= LintRemCancel
				}
				l.Ve = 0 // no pragmas in remark allowed
			}
		case inName: // split name on dots and spaces
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
			case isStructure(c):
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
							b = nil           // don't hold to backing array
							if ok := oc.linePragmas.lpcall[n](c, oc, oc.linePragmas.lpfpar[n]); !ok {
								oc.BadLint = OcLint{ln, LintBadLnPrag}
								return false
							}
							b = oc.Inbuf[:]
							blen = len(b)
							switch {
							case blen <= p || blen >= u32max:
								fallthrough
							default:
								// if handler messed up, we can not reliably proceed
								oc.BadLint = OcLint{ln, LintBufCorrupt}
								return false
							case oc.Inpos > p && b[oc.Inpos] == '\n': // modified OK
								p = oc.Inpos - 1
								continue
							case blen == blenwas && oc.Inpos == p && b[oc.Inpos] == c:
								break
							}
							break
						} // got line pragma to call
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
			case c < 0x20, blen-p < 4: // got empty value
				l.Vs = uint32(p + 1) // blen-p: 3210  43210  543210
				l.Ve = uint32(p + 1) // buffer:  :⬩$   : ⬩$   : .⬩$
				break                //                       := S$
			case c == '=' && b[p+2] == '=' &&
				b[p+3] < 0x21 && oc.AllowRaw: // here blen-p >= 4
				gotRaw = true
				l.Vs = uint32(p + 1)
				break
			case c == 0x20,
				c == ':' && b[p+2] == ' ':
				l.Vs = uint32(p + 2)
				break
			default:
				nowStage = fromStage
				continue // not a separator
			}
			if nowStage == lpCheck { // ORD item
				l.Ne = uint32(p)
				l.Fl |= IsOrd
			} else { // NAV item
				l.Ne = uint32(lastP + 1)
			}
			if !gotQuote && l.Ne > 0 && isStructure(b[l.Ne-1]) {
				l.Fl |= IsSpec
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
					oc.LapsesFound++ // take note
					if lint {
						lapses = append(lapses, OcLint{ln, culint}) // store
					}
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
					if r, ok := metaCheck(b, l.Vs, i); ok {
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
					if c != ' ' && !isPragmaNotMeta(c) { // no meta here
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
					if i < l.Vs && isPragmaChar(c) { // even lone pragma
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
			if gotRaw {
				gotRaw = false
				if l.Vs != l.Ve && l.Ve-l.Vs > 10 {
					for i := uint32(3); i < 11; i++ {
						rawB <<= 8
						rawB |= uint64(b[l.Vs+i])
					}
				} else {
					rawB = rawBoundary
				}
				var x uint64
				g := p + 1 // b[p]==0x10 ???
				for g < blen {
					c = b[g]
					switch {
					case c == 0x0a:
						ln++
					case c == 0x0d:
					case c < rawThre:
						oc.LapsesFound++
						oc.BadLint = OcLint{ln, LintCtlChars}
						return false
					}
					x <<= 8
					x |= uint64(c)
					if x == rawB {
						g -= 7
						break
					}
					g++
				}
				if blen-g < 8 { // no boundary found, FATAL
					oc.LapsesFound++
					oc.BadLint = OcLint{ln, LintNoBoundary}
					return false
				}
				l.Vs = uint32(p) + 1
				l.Ve = uint32(g)
				for g < blen { // move to the next line
					if b[g] == 0x0a {
						ln++
						p = g
						break
					}
					g++
				}
			} // if gotRaw block
			if culint != 0 { // store linted
				oc.LapsesFound++ // take note
				if lint {
					lapses = append(lapses, OcLint{ln, culint})
				}
				culint = 0
			}
			items = append(items, l) // store item
			l = OcItem{}
			ln++
			// default:
			//	culint |= LintUnknown // badChar check instead
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
} // func (oc *Flat) Tokenize (ok bool)

// func isStructure checks if c is an Oconf's structure bracket.
// This function is supposed to be inlined by the compiler.
func isStructure(c byte) bool { // ^ @ () [] {} <>
	return c == '^' || c == '@' || // section, model section
		c == '(' || c == ')' || // group
		c == '[' || c == ']' || // list
		c == '{' || c == '}' || // dict
		c == '<' || c == '>' //    set
	// 10 tests to fail
}

// func isPragmaChar checks if c is a pragma including meta brackets.
// This function is supposed to be inlined by the compiler.
func isPragmaChar(c byte) bool {
	return c == 0x27 || //          '
		(c > 0x28 && c < 0x2e) || //  ) * + , -
		(c > 0x21 && c < 0x26) || //  " # $ %
		c == 0x3b || c == 0x3e || //  ; >
		(c > 0x5b && c < 0x61) || //  \ ] ^ _ `
		(c > 0x7b && c < 0x7f) || //  | } ~
		c == 0x3f || c == 0x2f //     ? /
	// 13 tests to fail
}

// func isPragmaNotMeta checks if c is a only a pragma character.
// This function is supposed to be inlined by the compiler.
func isPragmaNotMeta(c byte) bool {
	return c == 0x27 || //          '
		c == 0x7c || c == 0x5c || //  | \
		(c > 0x5d && c < 0x61) || //  ^ _ `
		(c > 0x29 && c < 0x2e) || //  * + , -
		(c > 0x21 && c < 0x26) || //  " # $ %
		c == 0x3f || c == 0x7e //     ? ~
	// 11 tests to fail
}

// func metaCheck returns ok and r pointing to the meta start position;
// or !ok if stop position was reached before meta opening character
// was found. It allows for chained metas, even of different kind.
func metaCheck(b []byte, stop, i uint32) (r uint32, ok bool) {
	var c, d, e, o byte
	r = i
again:
	o = b[i]
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
	default:
		return
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

// func RegisterLinePragma adds user defined line pragma handler.
// Up to 8 different line pragmas can be registered simultaneously.
func RegisterLinePragma(lpc byte, oc *OcFlat,
	lpfunc LpHandler, lpfpar interface{}) (ok bool) {
	ld := &oc.linePragmas
	i := ld.lpchar
	if lpc > 0x2f || i&(255<<56) != 0 {
		return // bad char or all slots used.
	}
	var p byte
	for ; i != 0; i >>= 8 {
		if lpc == byte(i&255) {
			return // already set.
		}
		p++
	}
	ld.lpcall[p] = lpfunc
	ld.lpfpar[p] = lpfpar
	ld.lpchar |= lpchrint(lpc) << (p << 3)
	return true
}
