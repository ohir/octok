## Octok OCONF (object config) format tokenizer in Go.

`import "github.com/ohir/octok"`

### Overview

This package provides tokenizer for OCONF line format. Oconf config line 
format, while decades old, for its purposes still beats now ubiquotus YAML, JSON,
and even TOML. Both in readability for humans and for compactness of the parser.
OCONF simply delivers on all unfulfilled promises of YAML.

Oconf uses colon as key/value separator. Instead of quoting and escaping every
key and every value, OCONF uses a few digraphs, called "pragmas", put after the
value string if -- and only **if** -- the Value needs some special treatment. 
Like **`\t`**  unescaping. See below.


### Sample config, annotated. (Copied from OCONF's spec):

```
// line comment can also begin with " # ! (d-quote, sharp and bang) markers.
 ! free comment lines are possible too - these may not contain ' : ' though.

 ^ Section : ----- section lead --- //  ^  is a "section" marker and depth indicator. 
 '^ escape : not a section lead     // '   at start makes any key valid and ordinary.
    spaced :  val & spaces     |.   //  |  "guard pragma" keeps tail blanks intact.
    noComm : hello // there    '.   //  '  "disa  pragma" makes former // to the Value
   withCTL : Use\t tab and \n  \.   //  \  "unesc pragma" unescapes \t, \n, and \xHH
    withNL : some value        ^.   //  ^  "newline pragma" adds a '\n' at the end.
    looong : value can span    +.   //  +  "join Pragma" joins this line value with
           :  many lines and   +.   //      next line value 
           :: still keep indent.    // ::  separator makes leading space more visible.

  ^^ SubSec : --------------------- // ^^  open subsection at depth 2 (^^ or ^2 or ^).
            : list member  0        //     Ordered (unnamed) values can be indexed 
            : list member  1        //     naturally by the order of apperance
        33  : list member 33        //     or with index being given explicit
            : value                 //     /Section/SubSect[34] = value
                                    // 
   ^^^ SSSub : -------------------- // ^^^ go depth 3 sub-sub section
         key : value                //     /Section/SubSect/SSSub.key = value

 ^ OthSect : ---------------------- // Next depth 1 section opens. All above close.
       key : value                  // /OthSect.key = value
     a key : value                  // spaces in keys are ok.   Here is  "a key".
   ' spkey :  value                 // '  quotes leading space. Here is " spkey".
       Имя : Юрий                   // OConf supports utf-8 encoded unicode
       键k : v值                    // in full range. [And 8bit "codepages" too].

 ^^ SubDict : --------------------- // Show other structure constructs. These 
   'nodict { :                      // ' makes key ordinary "nodict {" 
  dictname { : dict opens           // SHOULD NOT be used in human editable
        some : value                // configs.
    listname [ : list opens         // Ordered (unnamed) values can be indexed 
               : list member 0      // naturally by the order of apperance
           33  : list member 33     // or with index being given explicit
               < : anon set opens   // <set> is now at index 34
                 : with unnamed    
             and : with named members 
        deepdict { : even with a dict 
              deep : value here  // /OthSect/SubDict/listname/34/deepdict/deep/
                 } : deep dict closes
               > : set closes
               : list member 35
             ] : list closes
       other : value
           } : dict closes

 ^^ PGroups : --------------------- // :( Group applies a pragma to many items
   tx1 ( group pragma ^+.           // Put ^+. on every line till group ends.
       : many lines may come here
       :  that keep indent line but
       :  sometimes need to be disa
       : mbiguated for // or ?.     '.
       ) group ends

  tx2 :== xHereRaw // block of raw text follows.             
       Now multiline text can span many lines. It ends at the boundary
       string that is provided after the :== marker. Custom boundary
       string must have at least 8 bytes and only 8 bytes of it matters.
       If custom is not given, boundary string defaults to ==RawEnd.
       Now, anything from x of xHereRaw to the end of line is a comment.
```

[More about OCONF](https://github.com/ohir/oconf-std)

### Package Documentation

[Documentation](http://godoc.org/github.com/ohir/octok) is hosted at GoDoc project.


### Install

Install package:

```
% go get -u github.com/ohir/octok
```

Then make symbols for tests:
```
% go get -u golang.org/x/tools/cmd/stringer 
% cd $GOPATH/github.com/ohir/octok
% go generate

% go test # should pass ok
```

### Revisions

  - v1.0.0 - public release
  - noversion, based on an old C/perl code.


### License

MIT. See LICENSE file.

(c)2019 Wojciech S. Czarnecki, OHIR-RIPE

