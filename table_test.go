package octok

// TESTING.
// Oconf tokenizer does bytes. As per utf8 basics, encoding is transparent
// to the code. Nonetheless, I decided to make users of this package happy
// with easy testing in their native scripts. Enjoy.
//
// You MUST prepare a failing test before raising a github Issue with
// octok.  Such test table entry should accompany any configuration line
// (or block of lines) you attached as one that parsed erroneously.
//
// Test and 'mock' strings given need to be byte to byte synchronized.
// It is done using markers and fillers of same byte- and wide-lengths.
// For wide print (asian) test glyphs mock uses ones from the 3w column.
//
// Mock markers and fillers:
// 0000 Hex is a codepoint in U+
//            1   2           3          3w          4 bytes
//   filler       · 00B7 b7   ‾ 203E be  ＿ FF3F bf  𝜋 1D70B 8b
//   Ns       N   И 0418 98   ↦ 21A6 a6  Ｎ FF2E ae  𝑁 1D441 81
//   Ne       n   и 0438 b8   ↤ 21A4 a4  ｎ FF4e 8e  𝑛 1D45B 9b
//   Vs       V   Ш 0428 a8   ↳ 21B3 b3  Ｖ FF3F b6  𝑉 1D449 89
//   Ve       v   ш 0448 88   ↲ 21B2 b2  ｖ FF56 96  𝑣 1D463 a3
//                (cyrrylic     arrows   full width  math alnum)
//
//   sep   :  - separator should be given in mock too
//   empty E  - mark empty value position in longer line. Eg. ": E   ^"
//   Ps    P  - Pragma start position
//   Ms    M  - meta start position
//   Pe    ^  - Next item begins here. Ie. tested one ends right before.
//         §  - Pe marker used only under test's CRLF (ø).
//               ^ or § are set right after last byte of the item tested.
//
// Test string newlines are marked with $ or ^ for NL, and ø for CRLF.
// Mock string uses either ^ or § to mark end of this/start of next item.
// $ is the default, replacing ^ for NL must be turned on via description
// string at first or second position:  `@^ description string`.
// The ! and & in mock are replaced by the TAB and DEL control chars.
//
// Test controls:
// Desc: Describe test. First char may instruct testing:
//       @ - this is the last test to perform.
//       ! - skip this one test.
//       ^ - use ^ as NL mark instead of $. Can be given after @ too.
// pRes: ParseNone - no parse result is expected. Errs if anything parsed.
//       ParseOK   - check first and only item.
//       Parsed2   - check second item of 2 to parse.
//       Parsed3   - check second item of 3 to parse.
//
//23456789¹123456789²123456789³123456789⁴123456789⁵123456789⁶123456789⁷1234567
// C&P markers/fillers:  · И и Ш ш : ‾ ↦ ↤ ↳ ↲ : ＿ Ｎ ｎ Ｖ ｖ : 𝜋 𝑁  𝑛 𝑉 𝑣
// change flag +/- 1 to fail and you can see how a given test string parsed
var ttableFirstItem = 55        // sync it for accurate test error messages
var tokTestTable = []ptestItem{ // desc, Fl, Tt, Np, pResult / from / mock
	{`^Utf8 Supermix with two rems and two metas`, NextCont | Backtick, 0x00, 0x0000, LintOK, Parsed2,
		"k : v //rem ^  🁩λϕ🂊🁾Σbuξ℉⇶av⸨ : ⸩𐌓ar𐌕ó𐌁𐌅兄 `+@内ꜸԳՔЖ№;<żó丶ć>. //remarkø",
		"               𝑁··𝜋𝜋·  ·‾‾  ↤ : ↳𝜋  𝜋·𝜋𝜋ｖ P M＿‾···‾  ··＿·  ^        ·"},
	{`Utf8 Supermix with rem`, NextCont | Backtick, 0x00, 0x0000, LintOK, ParseOK,
		"🁩λϕ🂊🁾Σbuξ℉⇶av⸨ : ⸩𐌓ar𐌕ó𐌁𐌅兄 `+@内ꜸԳՔЖ№;. //remarkø",
		"𝑁··𝜋𝜋·  ·‾‾  ↤ : ↳𝜋  𝜋·𝜋𝜋ｖ P M＿‾···‾  ^        ·"},
	{`Utf8 Supermix`, NextCont | Backtick, 0x00, 0x0000, LintOK, ParseOK,
		"🁩λϕ🂊🁾Σbuξ℉⇶av⸨ : ⸩𐌓ar𐌕ó𐌁𐌅兄 `+@内ꜸԳՔЖ№;.ø",
		"𝑁··𝜋𝜋·  ·‾‾  ↤ : ↳𝜋  𝜋·𝜋𝜋ｖ P M＿‾···‾  §"},
	{`Utf8 4B`, NextCont | Backtick, 0x00, 0x0025, LintOK, ParseOK,
		"𝌆 𝌢  : 𝍍 𝌌 𝌶 `+[🁩 🂊 🁾].$",
		"𝑁 𝑛  : 𝑉 𝜋 𝑣 P M𝜋 𝜋 𝜋  ^"},
	{`Chinese keyparts pragma 3B`, NextCont, 0x3f, 0x0027, LintOK, ParseOK,
		"更长 的钥匙 : 更长的价值 ?+(元数据).$",
		"Ｎ＿ ＿＿ｎ : Ｖ＿＿＿ｖ P M＿＿＿  ^"},
	{`Chinese simple 3B`, NextCont | Backtick, 0x00, 0x0000, LintOK, ParseOK,
		"键 : 值 `+[元標記].$",
		"ｎ : ｖ P M＿＿＿  ^"},
	{`Cyryllic compound 2B w/ path and rem`, NextCont, 0x3f, 0x0027, LintOK, ParseOK,
		"   Имя Юрий : Шнурок ?+=/path/here/. //rem here$",
		"   И·· ···и : Ш····ш P M            ^           "},
	{`Cyryllic compound 2B`, NextCont, 0x3f, 0x0027, LintOK, ParseOK,
		"Имя Юрий : Шнурок ?+<tag>.$",
		"И·· ···и : Ш····ш P M     ^"},
	{`Some cyrillic 2B`, NextCont, 0x3f, 0x0000, LintOK, ParseOK,
		"Имя : Шнурок ?+<tag>.$",
		"И·и : Ш····ш P M     ^"},
	{`CRLF simple`, NextCont, 0x3f, 0x0000, LintOK, ParseOK,
		"name : value ?+<tag>.ø",
		"N  n : V   v P M     §"},
	{`use filler`, Unescape | Backtick, 0x83, 0x0000, LintOK, ParseOK,
		"name : val '_\\__^^^_`_(meta).$",
		"N  n : V v P          M      ^"}, // \\ shift-N
	{`guard repeated`, NoneF, 0x00, 0x0000, LintSusPragma, ParseOK,
		"name : value ||<tag>.$",
		"N  n : V            v^"},
	{`dot parted key`, NoneF, 0x00, 0x04a9, LintOK, ParseOK,
		"Some.Key.Here : value$",
		"N           n   V   v^"},
	{`too many part name`, NoneF, 0x00, 0x8cca, LintKeyParts, ParseOK,
		"na me na2 me2 na3 me3 : v$",
		"N                   n : v^"},
	{`no parts name from the dot`, NoneF, 0x00, 0x0000, LintOK, ParseOK,
		" '.name : v$",
		"  N   n : v^"},
	{`three parts name from the space`, NoneF, 0x00, 0x04ca, LintOK, ParseOK,
		" ' name two.three : v$",
		"  N             n : v^"},
	{`three parts name from many spaces`, NoneF, 0x00, 0xa1b5, LintOK, ParseOK,
		" '   name two  three   four : v$",
		"  N                       n : v^"},
	{`four parts name from the dot`, NoneF, 0x00, 0x9950, LintOK, ParseOK,
		" '.name.two.three.four : v$",
		"  N                  n : v^"},
	{`five parts name from the dot`, NoneF, 0x00, 0x9950, LintKeyParts, ParseOK,
		" '.name.two.three.four.five : v$",
		"  N                       n : v^"},
	{`dot parted key dot lead`, NoneF, 0x00, 0x04ca, LintOK, ParseOK,
		"'.Some.Key.Here : value$",
		" N            n   V   v^"},
	{`dot parted key, last part empty`, NoneF, 0x00, 0x952e, LintOK, ParseOK,
		"Some.Key.Here. : value$",
		"N            n   V   v^"},
	{`dot and space parted key`, NoneF, 0x00, 0x952e, LintOK, ParseOK,
		"Some Key.Here Is : value$",
		"N              n   V   v^"},
	{`three part name`, NoneF, 0x00, 0x0466, LintOK, ParseOK,
		"na me 3 : v$",
		"N     n   v^"},
	{`name starts with colon, val empty`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"':a :$",
		" Nn :^"},
	{`single colon name, val`, NoneF, 0x00, 0x0000, LintOK, ParseOK,
		"': : v$",
		" n : v^"},
	{`single colon name, val empty`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"': :$",
		" n :^"},
	{`single colon name, val empty2`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"':   : $",
		" n   :^ "},
	{`single colon name, val empty3`, IsEmpty, 0x00, 0x0000, LintBadEndLin, ParseNone,
		"':   :      $ asasa ",
		" n   : ^            "},
	{`shortest ord + no NL tail`, IsOrd | IsEmpty, 0x00, 0x0000, LintBadEndLin, ParseNone,
		":$  name : value     ",
		":^                   "},
	{`Many spaces after possible pragma`, IsOrd, 0x00, 0x0000, LintOK, ParseOK,
		": v. ?.            .$",
		": V                v^"},
	{`single colon empty ord`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		":$",
		":^"},
	{`no space at ord pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseNone,
		":'.$",
		":P ^"},
	{`no space at named pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseNone,
		"n :'.$",
		":  P ^"},
	{`empty ord disambiguated`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": '.$",
		": P ^"},
	{`empty ord guarded`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": |.$",
		": P ^"},
	{`empty named guarded`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"n : |.$",
		"n : P ^"},
	{`empty named guarded w/meta`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"n : |<tag>.$",
		"n : PM     ^"},
	{`ord single space`, IsOrd, 0x00, 0x0000, LintOK, ParseOK,
		":: |.$",
		": vP ^"},
	{`named empty pragonly`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"    n : '.$",
		"    n : P ^"},
	{`named empty metaonly`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"     n : [mmme].$",
		"     n : M      ^"},
	{`named empty metaonly2`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"   n :   [me].$",
		"   n : E M    ^"},
	{`named many metas only`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"   n :   [me]<tag>@type1;@type2;@type3;(group too).$",
		"   n : E M                                         ^"},
	{`named empty disambiguated`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		" n : '.$",
		" n : P ^"},
	{`single dot is not a pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": // x.$",
		": ^     "},
	{`last pragma disambiguates also rem`, IsOrd, 0x3f, 0x0000, LintRemCancel, ParseOK,
		":  // ?.$",
		": V v P ^"},
	{`norem should be disambiguated explicit`, IsOrd, 0x00, 0x0000, LintOK, ParseOK,
		":  // '.$",
		": V v P ^"},
	{`any pragma disambiguates too`, IsOrd, 0x3f, 0x0000, LintOK, ParseOK,
		": +. ?.$",
		": Vv P ^"},
	{`remkey empt w/pragma distant bad rem`, NoneF, 0x00, 0x0023, LintOK, ParseOK,
		"Na //me :   +.// norem  $",
		"N     n : V          v^  "},
	{`disambiguation after rem`, NoneF, 0x00, 0x0023, LintOK, ParseOK,
		"Na //me :   +. // rem:. '.$", // disambiguate
		"N     n : V           v P ^"},
	{`remkey empt w/pragma distant and broken rem`, NoneF, 0x00, 0x0023, LintSusPragma | LintRemCancel, ParseOK,
		"Na //me :   +. // rem   ;. $", // pragma-like can not be in a remark
		"N     n : V              v^ "},
	{`remkey empt w/pragma distant and rem`, IsEmpty | NextCont, 0x00, 0x0023, LintOK, ParseOK,
		"Na //me :   +. // rem  $",
		"N     n : E P ^         "},
	{`remkey empty w/pragma close and rem`, IsEmpty | NextCont, 0x00, 0x0023, LintOK, ParseOK,
		"Na //me : +. // rem  $",
		"N     n : P ^         "},
	{`no remark valid in name`, IsEmpty, 0x00, 0x0023, LintOK, ParseOK,
		"Na //me :   // remark  $",
		"N     n : ^             "},
	{`remark then spaces to EoF`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"Name :   // remark  $        ",
		"N  n : ^                     "},
	{`remark then EoF comment`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"Name :   // remark  $  # End comment no NL ",
		"N  n : ^                                   "},
	{`remark then something to no NL EoF`, IsEmpty, 0x00, 0x0000, LintBadEndLin, ParseNone,
		"Name :   // remark  $  Something more no NL ",
		"N  n : ^                                    "},
	{`rem right after empty value`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"     n : //[mmme]  $",
		"     n : ^          "},
	{`skip empty rem right after empty value`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"$  $  $  n : //[mmme]  $",
		"         n : ^          "},
	{`del in line`, NoneF, 0x00, 0x0000, LintCtlChars, ParseOK,
		" n&n : v$ n : v$",
		"          n : v^"},
	{`tab in line`, NoneF, 0x00, 0x0000, LintCtlChars, ParseOK,
		" n!n : v$ n : v$",
		"          n : v^"},
	{`zero in line`, NoneF, 0x00, 0x0000, LintCtlChars, ParseOK,
		" n\000n : v$ n : v$",
		"          n : v^"},
	{`apo single`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		" ' : $",
		" ' :^ "},
	{`apo bad`, NoneF, 0x00, 0x0000, LintOK, ParseNone,
		"':$",
		" :^"},
	{`apo here`, IsOrd | IsEmpty, 0x00, 0x0000, LintNoComment, ParseOK, // lint: "'$ part
		"'$ ' :$",
		"   ' :^"},
	{`apo skipped`, IsOrd | IsEmpty, 0x00, 0x0000, LintNoComment, ParseOK, // lint: "'$ part
		"'$:$",
		"  :^"},
	{`apo empty`, IsOrd | IsEmpty, 0x00, 0x0000, LintNoComment, ParseOK, // lint: " ' $ part
		" ' $:$",
		"    :^"},
	{`apo end`, IsOrd | IsEmpty, 0x00, 0x0000, LintNoComment, ParseOK, // lint: " '$ part
		" '$: $",
		"   :^ "},
	{`empty ord 1`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"  : $",
		"  :^ "},
	{`empty ord 2`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"  :      $",
		"  : ^     "},
	{`short2 ord`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		":   $",
		": ^  "},
	{`short1 ord`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		":  $",
		": ^ "},
	{`shortest ord`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		":$",
		":^"},
	{`empty ord with many metas`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"     :   [me]<tag>@type 1;@type 2;@type 3;{xyz}(group too).$",
		"     : E M                                                 ^"},
	{`indexed ord with many metas`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		" 276 :   [me]<tag>@type 1;@type 2;@type 3;{xyz}(group too).$",
		" N n : E M                                                 ^"},
	{`short value`, IsOrd, 0x00, 0x0000, LintOK, ParseOK,
		"  : v    $",
		"  : v^    "},
	{`short name/value`, NoneF, 0x00, 0x0000, LintOK, ParseOK,
		"  n : v$",
		"  n : v^"},
	{`simple name`, NoneF, 0x00, 0x0000, LintOK, ParseOK,
		"Name : Value$",
		"N  n : V   v^"},
	{`type and newline pragma`, NoneF, 0xc3, 0x0000, LintSusPragma | LintTypeAndNL, ParseOK,
		"Name : Value ?^.  $",
		"N  n : V       v^  "},
	{`join +% together, and ^`, NextCont | NextMeta, 0x81, 0x0000, LintTwoJoins, ParseOK,
		"Name : Value +%^.  $",
		"N  n : V   v P   ^  "},
	{`join +% together, and ^`, NextCont | NextMeta, 0x81, 0x0000, LintTwoJoins, ParseOK,
		"Name : Value %+^.  $",
		"N  n : V   v P   ^  "},
	{`type and ^`, NoneF, 0xc3, 0x0000, LintSusPragma | LintTypeAndNL, ParseOK,
		"Name : Value ?^.  $",
		"N  n : V       v^  "},
	{`newline and type pragma`, NoneF, 0x81, 0x0000, LintTypeAndNL, ParseOK,
		"Name : Value ^?.  $",
		"N  n : V   v P  ^  "},
	{`newline pragma`, NextCont, 0x81, 0x0000, LintOK, ParseOK,
		"Name : Value ^+.  $",
		"N  n : V   v P  ^  "},
	{`many newlines pragma`, NextCont, 0x85, 0x0000, LintOK, ParseOK,
		"Name : Value ^^^^^+.  $",
		"N  n : V   v P      ^  "},
	{`broken newlines pragma`, NextCont, 0x83, 0x0000, LintDublCaret, ParseOK,
		"Name : Value ^^^+^^.  $",
		"N  n : V   v P      ^  "},
	{`too many (65) newlines pragma`, NextCont, 0xc1, 0x0000, LintSusPragma | LintTooManyNL, ParseOK,
		"Name : Value ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^+.$",
		"N  n : V                                                                       v^"},
	{`simple pragma`, NextCont, 0x00, 0x0000, LintOK, ParseOK,
		"Name : Value +.  $",
		"N  n : V   v P ^  "},
	{`simple <>meta`, NoneF, 0x00, 0x0000, LintOK, ParseOK,
		"Name : Value <meta>.   $",
		"N  n : V   v M      ^   "},
	{`simple pragma+meta`, NextMeta, 0x00, 0x0000, LintOK, ParseOK,
		"Name : Value %<meta>.   $",
		"N  n : V   v PM      ^   "},
	{`simple pragma+meta crlf`, NextMeta, 0x00, 0x0000, LintOK, ParseOK,
		"Name : Value %<meta>.ø",
		"N  n : V   v PM      §"},
	{`value like a next meta`, NoneF, 0x00, 0x0000, LintOK, ParseOK,
		"Name : <Value> [meta].   $",
		"N  n : V     v M      ^   "},
	{`suspected pragma/broken meta`, NoneF, 0x00, 0x0000, LintSusPragma, ParseOK,
		"Name :   Value >[meta].   $",
		"N  n : V              v^   "},
	{`types two types together "`, IsOrd, 0xc2, 0x0000, LintSusPragma | LintManyTypes, ParseOK,
		`: ?".$`,
		`: V v^`},
	{`types type "`, IsOrd | IsEmpty, 0x22, 0x0000, LintOK, ParseOK,
		`: ".$`,
		`: P ^`},
	{`types type ?`, IsOrd | IsEmpty, 0x3f, 0x0000, LintOK, ParseOK,
		": ?.$",
		": P ^"},
	{`types type #`, IsOrd | IsEmpty, 0x23, 0x0000, LintOK, ParseOK,
		": #.$",
		": P ^"},
	{`^ypes type $`, IsOrd | IsEmpty, 0x24, 0x0000, LintOK, ParseOK,
		": $.^",
		": P ^"},
	{`types type ,`, IsOrd | IsEmpty, 0x2c, 0x0000, LintOK, ParseOK,
		": ,.$",
		": P ^"},
	{`types type ~`, IsOrd | IsEmpty, 0x7e, 0x0000, LintOK, ParseOK,
		": ~.$",
		": P ^"},
	{`types type *`, IsOrd | IsEmpty, 0x2a, 0x0000, LintOK, ParseOK,
		": *.$",
		": P ^"},
	{`types type -`, IsOrd | IsEmpty, 0x2d, 0x0000, LintOK, ParseOK,
		": -.$",
		": P ^"},
	{`succesful % line pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": '.$% line pragma$",
		": P ^              "},
	{`^succesful $ line pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": '.^$ line pragma^",
		": P ^              "},
	{`succesful + line pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": '.$+ line pragma$",
		": P ^              "},
	{`succesful , line pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": '.$,  line pragma$",
		": P ^               "},
	{`succesful - line pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": '.$- line pragma$",
		": P ^              "},
	{`succesful . line pragma`, IsOrd | IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		": '.$. line pragma$",
		": P ^              "},
	{`Section 1`, IsEmpty | IsSpec, 0x00, 0x0029, LintOK, ParseOK,
		"Section1 ^ :$",
		"N        n :^"},
	{`Section 3 space`, IsSpec, 0x00, 0x048d, LintOK, ParseOK,
		"[[[ Section3 ]]] : val $",
		"N              n   V v^ "},
	{`Section 3 dot`, IsSpec, 0x00, 0x048d, LintOK, ParseOK,
		"[[[.Section3.]]] : val $",
		"N              n   V v^ "},
	{`Section 3 lead`, IsSpec, 0x00, 0x0024, LintOK, ParseOK,
		"[[[ Section3   : val $",
		"N          n     V v^ "},
	{`Section in the middle`, IsSpec, 0x00, 0x052c, LintOK, ParseOK,
		"Section2 ^^ xy : val $",
		"N            n : V v^ "},
	{`Section 1 lead`, IsEmpty | IsSpec, 0x00, 0x0000, LintOK, ParseOK,
		"^Section1 :$",
		"N       n :^"},
	{`Section 5 lead`, IsEmpty | IsSpec, 0x00, 0x0026, LintOK, ParseOK,
		"^^^^^ Section5 :$",
		"N            n :^"},
	{`List start named`, IsEmpty | IsSpec, 0x00, 0x0466, LintOK, ParseOK,
		"na me [ :$",
		"N     n  ^"},
	{`List start ord`, IsOrd | IsEmpty | IsSpec, 0x00, 0x0022, LintOK, ParseOK,
		"5 [ :$",
		"N n :^"},
	{`List start`, IsEmpty | IsSpec, 0x00, 0x0000, LintOK, ParseOK,
		"[ :$",
		"n :^"},
	{`List end`, IsEmpty | IsSpec, 0x00, 0x0000, LintOK, ParseOK,
		"] : $",
		"n  ^ "},
	{`Dict start named`, IsEmpty | IsSpec, 0x00, 0x0466, LintOK, ParseOK,
		"na me { :$",
		"N     n  ^"},
	{`Dict end`, IsEmpty | IsSpec, 0x00, 0x0000, LintOK, ParseOK,
		"} :$",
		"n  ^"},
	{`Group start`, NextCont | IsSpec, 0x00, 0x0466, LintOK, ParseOK,
		"na me ( : new group '+.  $",
		"N     n   V       v P  ^  "},
	{`Group end`, IsEmpty | IsSpec, 0x00, 0x0000, LintOK, ParseOK,
		") :$",
		"n  ^"},
	{`Set start`, IsEmpty | IsSpec, 0x00, 0x0466, LintOK, ParseOK,
		"na me < :$",
		"N     n  ^"},
	{`Set not start`, IsEmpty, 0x00, 0x0023, LintOK, ParseOK,
		"na me< :$",
		"N    n :^"},
	{`Set end`, IsEmpty | IsSpec, 0x00, 0x0000, LintOK, ParseOK,
		" > :$",
		" n  ^"},
	{`Set end`, IsEmpty | IsSpec, 0x00, 0x0000, LintOK, ParseOK,
		"> :$",
		"n  ^"},
	{`escaped Section 1`, IsEmpty, 0x00, 0x0000, LintOK, ParseOK,
		"'>Section1 :$",
		"'N       n :^"},
	{`escaped Section 5`, IsEmpty, 0x00, 0x0026, LintOK, ParseOK,
		"'>>>>> Section5 :$",
		"'N            n :^"},
}

type ptestItem struct {
	desc string   // description
	fl   ItemFL   // flags
	tt   byte     // type char/nlines count
	np   uint16   // name parts
	lint LintFL   // linter output
	pres paResult // parser result
	from string   // to parse
	mock string   // result string describing expected positions:
}
