package octok

// OcFlat keeps the raw text buffer and after Tokenize() it has its
// []Items filled with parsed Item offsets pointing into that buffer.
// OcFlat typically is embedded in a some "Parser" or "Config" struct.
type OcFlat struct {
	Inbuf         []byte     // raw input buffer
	Items         []OcItem   // parsed lines
	Lapses        []OcLint   // lints found. Filled if LintFull is true.
	BadLint       OcLint     // why !ok
	Inpos         int        // parser position - updated on line pragma calls only.
	InLine        uint32     // pragma call line
	ItemsExpected uint32     // default 64
	LapsesFound   uint32     // lints counter, incemented even if LintFull is false
	LintFull      bool       // register lints. Otherwise just up LapsesFound.
	AllowBinRaw   bool       // allow binary raw values, otherwise just \t\n\r\del
	NoTypes       bool       // disallow all type chars: - * ~ , $ # ? "
	NoMetas       bool       // disallow all metas:  @…; =…/ (…) […] {…} <…>
	Pck           uint64     // reference linter must be configurable
	Sck           uint64     // 32B wasted for production code, an hour
	Mck           uint64     // of time NOT dealing with separate "for
	Tck           uint64     // linter" type saved; per every person.
	linePragmas   lpDispatch // registered line pragma handlers
}

// OcItem keeps an oconf's ITEM found within Inbuf by Tokenize().
//
// Tc and Np fields of the OcItem are bit packed:
//
// 		Tc field contains either type character (if b7==0)
// 		or number of pragma carets (max^ is 63). Therefore:
//
// 		(Tc.b7 == 1) is effectively the 'HasCarets' flag.
// 		(Tc != 0 && Tc < 128) makes an 'IsTyped' flag.
//
// Np field contains up to three 5-bit offsets to the name parts,
// so reference implementation recognizes up to four key parts.
// (First part spans from Ns to the first offset in Np. Next three parts
// start at further Ns+Np.offsets.)
//
// 5bit means also that last recognized part of the name must start no
// further than 31 bytes from the Ns position.
// It is not an error if there are more or further placed parts but those
// will not be registered in Np. (Linter marks this condition with
// 'LintKeyParts' flag.)
//
// The Ps field is a helper field. After Tokenize() it is free to
// be reused.
type OcItem struct {
	Ns uint32 // 4B Name  start position
	Ne uint32 //_4B Name end position
	Vs uint32 // 4B Value start position
	Ve uint32 //_4B Value end position
	Ps uint32 // 4B Pragma start position (used in tests).
	Ms uint32 //_4B Meta start position
	Pe uint32 // 4B Pragma end position. Item end position.
	Np uint16 // 2B Name parts 3x5b +1b flag
	Tc byte   // 1B type character or ^ counter. See below.
	Fl ItemFL //_1B flags
} // 32B

type OcItemNp = uint16

const NpOverParts OcItemNp = 1 << 15 // Np has all its parts filled

type OcItemTc = byte

const (
	// used with OcItem.Tc field
	TcHasCarets OcItemTc = 1 << 7 // Tc contains № of carets
	TcHasErrBit OcItemTc = 1 << 6 // for &testing
	TcHasError  OcItemTc = 0xc0   // for &testing
	TcTooManyNL OcItemTc = 0xc1   // Too many carets given
	TcDoublType OcItemTc = 0xc2   // Double type characters spotted
	TcTypeAndNL OcItemTc = 0xc3   // ^ and type chars in same pragma
)

//

// const SECT_LEAD keeps a character that is used as a Section marker.
const SectLead byte = '^'
const SectLeadEx byte = '@'

// parser stages
type pStage byte

const (
	lpCheck      pStage = 0 // order matters so better not use iota
	inName       pStage = 1
	ckSEP        pStage = 2
	inValue      pStage = 3
	registerItem pStage = 4
)

// do not change ItemFlags, test table got most of it as 0xXX.
// To be sure make manual not with (1 << iota) >> 1

// The OcItem.Fl (flags) field uses below constants:
type ItemFL byte

// ItemFlags conveys information about an Item. Take note, that Tokenize()
// does NOT recognize config's structure - it merely flags special names.
// Also dealing with explicit INDEX (all digits name) is left to the parser.
const ( // ItemFlags
	IsOrd    ItemFL = 1 << iota // ORD (ordered, not named) value.         " : value" item.
	IsEmpty                     // Value is empty.                 "name : " or " : " item.
	NextCont                    // +. pragma sets this
	NextMeta                    // %. pragma sets this.
	Unescape                    // \. pragma sets this.
	Backtick                    // `. pragma sets this.
	IsSpec                      // Name starts with a character of >[({})]< set
	IsIndex                     // Name starts with an ascii digit
	NoneF    ItemFL = 0         // Nothing special.         A straight "name : value" item.
)

// Linter recognized ambigous constructs are given as err-flags.
type LintFL uint32

const (
	LintSusPragma  LintFL = 1 << iota // Unconfirmed pragma spotted. Use '. next time.
	LintRemCancel                     // Pragma other than '. or |. cancelled endline remark.
	LintNoComment                     // Garbage text (or not marked comment) was skipped.
	LintDublCaret                     // Non consecutive carets given. These must be grouped.
	LintTooManyNL                     // More than 63 carets was seen in pragma.
	LintTypeAndNL                     // Type char and ^ was given in single pragma.
	LintTwoJoins                      // Join pragmas % and + given together. Thats impossible.
	LintManyTypes                     // More than one type character given.
	LintCtlChars                      // Ascii control characters spotted.
	LintBadLnPrag                     // Line pragma returned with error.
	LintBufCorrupt                    // Line pragma corrupted buffer. Can not proceed.
	LintKeyParts                      // Name has more than 4 parts or last part starts too far (>31).
	LintBadEndLin                     // No NL at the end of buffer. Last line had not registered.
	LintBadBufLen                     // Buffer is too short or too long to parse.
	LintNoBoundary                    // RawEnd boundary could NOT be found!
	LintUnknown                       // Test suite sentinel. Keep it at last entry.
	LintOK         LintFL = 0         // Linted OK.
)

// You need to manual update (f LintFL) Msg() method after adding a flag.
// It does not deserve automation.

// func LintMessage(flag LintFL) returns string, possibly multiline one,
// with linting results description.
func LintMessage(l LintFL) (r string) {
	var mt = [...]string{
		`Linted OK.`, // len:10
		`Unconfirmed pragma spotted. Use '. next time.`,
		`Pragma other than '. or |. cancelled endline remark.`,
		`Garbage text (or not marked comment) was skipped.`,
		`Non consecutive carets given. These must be grouped.`,
		`More than 63 carets was seen in pragma.`,
		`Type char and ^ was given in single pragma.`,
		`Join pragmas % and + given together. Thats impossible.`,
		`More than one type character given.`,
		`Ascii control characters spotted.`,
		`Line pragma returned with error.`,
		`Line pragma corrupted buffer. Can not proceed.`,
		`Name has more than 4 parts or last part starts too far (>31).`,
		`No NL at the end of buffer. Last line had not registered.`,
		`Buffer is too short or too long to parse.`,
		`RawEnd boundary could NOT be found!`,
		`Test suite sentinel. Keep it at last entry.`,
	}
	// yank constants, paste, vselect from 1 then: s#^.\+// #`# | '<,'>s#$#`,#
	// mind Linted OK position is at 0!

	if l == 0 {
		return mt[0]
	}
	nl := " ‣ "
	for i := 1; l != 0; l >>= 1 {
		if l&1 == 0 {
			i++
			continue
		}
		if i < len(mt) {
			r += nl + mt[i]
			nl = "\n ‣ "
		} else {
			return "Err!"
		}
		i++
	}
	return r
}

// Why not use stringer?
// $ stringer  -linecomment=true -output=msg_linter.go -type=LintFL
// then rename (l LintFL) String to LintMessage, BUT!
// But there is a years persisting heisenbug in stringer logic that often
// results in "stringer: no value for constant" error if constant
// is not declared in its own separate file. I won't fight with this.
// Also shift-loop can produce an "All" result in a single run. As seen
// above.

// OcLint keeps linter flags associated with an input buffer's line
// where the ambiguous construct was found.
type OcLint struct {
	Line uint32
	What LintFL
}

type lpchrint uint64
type lpDispatch struct {
	lpchar lpchrint
	lpcall [8]LpHandler   // pragma handler
	lpfpar [8]interface{} // handler parameters
}

// LpHandler func takes care of line pragmas. For the simplicity and power
// its in/out parameters are registered as a separate item of interface{}
// type. This way line pragma handler may affect the program
// state without wasting memory for convoluted closure over things that
// possibly might not exist at pre-configuration stage.
type LpCallParam = interface{}

// LpHandler func is called when Tokenize finds a registered line
// pragma.  At call time it is being given pragma character, the
// tokenizer object and a pointer to the user provided LpCallParam
// struct. LpHandler should return true on success.
//
//	type LpHandler func(pch byte, oc *OcFlat, fpar interface{}) (ok bool)
//
//	type myLpResult struct {
//		deployToTestEnv bool
//	}
//
//	var lpDollar myLpResult
//	func lpDollarHandler(pch byte, oc *OcFlat, fpar interface{}) (ok bool) {
//		if r, ok := fpar.(*myLpResult); ok { // concretize fpar
//			r.deployToTestEnv = true     // set some state
//			return true
//		}
//		return false
//	}
//	// ... Then, in some init func, register the handler:
//
//	var oc octok.OcFlat
//	if ok := octok.RegisterLinePragma('$', &oc, lpDollarHandler, &myLpResult); !ok {
//		// report registering error
//	}
//
// LpHandler has full access to the octok.OcFlat data and it is free to
// tinker with .Inbuf from the .Inpos position onward. Upon succesful
// return Tokenize() resumes parsing over possibly changed .Inbuf from the
// possibly changed .Inpos place. If buffer has changed, new .Inpos MUST
// point to a 0x0a (newline) byte, usually one that ends the very pragma
// line. If these conditions are not meet, Tokenize() immediately will
// return with !ok.
//
// LpHandler SHALL NOT touch already parsed things
//
// ie. neither OcFlat.Inbuf's part from the beginning to the passed Inpos
// nor the OcFlat.Items table!
//
type LpHandler func(pch byte, oc *OcFlat, fpar interface{}) (ok bool)

// This struct is used as a parameter to the LinterSetup. fine-tune the Linter.
// 		type LiPrCh = LinterPragmaChars // yet better alias it in your test code.
// P, T, M are used to shrink the sets (restrict to the ones given).
// S adds up to 8 characters treated as Special key markers. It may use
// any but letter character.
type LinterPragmaChars struct {
	P string // Pragma characters: _ ` % \ + ^ | '
	T string //   Type characters: - * ~ , $ # ? "
	M string //   Meta characters: } ) ] > / ; fill END chars only!
	S string //    Added Specials: none
}

// Characters recognized as per FULL oconf spec.
const (
	pragmaChars uint64 = 0x5f60255c2b5e7c27 // _ ` % \ + ^ | '
	typeChars   uint64 = 0x2d2a7e2c24233f22 // - * ~ , $ # ? "
	metaChars   uint64 = 0x00007d295d3e2f3b // } ) ] > / ;
	rawBoundary uint64 = 0x3d3d526177456e64 // ==RawEnd
)

const u32max = (1 << 32) - 1

/*
// OcItem compact version with Ns+ offsets [PRD-83-xx gear]
type OcItemC struct {
	Ns uint32 // 4B posit. Name start
	Ne byte   // 2B Name end. offset Ns+
	Tc byte   // 1B type/newlines
	Fl ItemFL // 1B flags
	Vs byte   // 2B offset Value start Ns+
	Ve uint16 // 2B offset Value end Ns+
	Ms uint16 // 2B offset Meta start
	Pe uint16 // 2B offset Pragma end Ns+. Gives Item end position.
} */ // 16B
