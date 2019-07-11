## Octok OCONF (object config) format tokenizer in Go.

`import "github.com/ohir/octok"`

### Overview

This package provides tokenizer for OCONF line format. Oconf config line 
format, while decades old, for its purposes still beats now ubiquotus YAML, JSON,
and even TOML. Both in readability for humans and for compactness of the parser.
OCONF simply delivers on all unfulfilled promises of YAML.

Oconf uses colon as key/value separator. Instead of quoting and escaping every
key and every value, OCONF uses a few digraphs, called "pragmas", put after the
value string if -- **and only if** -- the Value needs some special treatment. 
Like `\v` or `\x07` unescaping. See below.


### Sample config, annotated. (Copied from OCONF's spec):

```
// line comment can also begin with " # ! (d-quote, sharp and bang) markers.
 ! free comment lines are possible too - these may not contain ' : ' though.
 # Note that in real configs pragmas and annotations are very rare.

 ^ Section : ----- section lead --- //  ^  is a "section" marker and depth indicator. 
 '^ escape : not a section lead     // '   at start makes any key valid and ordinary.
    spaced :  val & spaces     |.   //  |  "guard pragma" keeps tail blanks intact.
    noComm : hello // there    '.   //  '  "disa  pragma" makes former // to the Value
   withCTL : Use\v vtab and \n \.   //  \  "unesc pragma" unescapes control characters
    withNL : some value        ^.   //  ^  "newline pragma" adds a '\n' at the end.
    looong : value can span    +.   //  +  "join Pragma" joins this line value with
           :  many lines and   +.   //      next line value 
           :: still keep indent.    // ::  separator makes leading space more visible.

  ^^ SubSec : --------------------- // ^^  open subsection at depth 2.
                                    //     ----- in value above is just a decoration
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

 ^^ PGroups : --------------------- // ( Group applies a pragma to many items
    ( : group pragma ^+.            // Put ^+. on every line till group ends.
      : many lines may come here    //  Eglible are metas and pragmas + \ ` ^. 
      :  that keep indent line but  //  | ' % can not be grouped. 
      :  sometimes need to be disa  //
      : mbiguated for // or ?.  '.  // Here sum of pragmas applies: '^+.
      : 
    ) : group ends                  // Bracket can have a value or pragma, too.

 ^^ SubTrees : -------------------- // Show other structure constructs.
  dictname { : dict opens           // These SHOULD NOT be used
        some : value                // in human __editable__ configs.
     'nodict { :                    // 'disa makes key ordinary: "nodict {" 
    listname [ : list opens         // Ordered (unnamed) values can be indexed 
               : list member 0      // naturally by the order of apperance
           33  : list member 33     // or with index being given explicit
               < : anon set opens   // <set> is now at index 34
                 : with unnamed    
             and : with named members 
            '33  : 33 is a string not an index due to opening 'disa
            ''7  : '7 is a two characters string
        deepdict { : dictionary in a set, looked up by its name
              deep : value here  // /OthSect/SubDict/listname/34/deepdict/deep/
                 } : deep dict closes
               > : set closes
               : list member 35
             ] : list closes
       other : value
           } : dict closes

 ^^ Raw Multiline :   // Use :==    
  mtx :== xHereRaw    // below multiline block will be a VALUE of mtx. 
      This^^^^^^^^ is a custom boundary string. It must have at least
      8 bytes and exactly 8 bytes of it matters. If no custom boundary
      given (or it is too short), the ==RawEnd default one is used.
      Here block ends at a space before the x of xHereRaw.
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

% go test -cover # should pass 100% ok
```

### Revisions

  - v0.3.0 - Fixed support for TAB whitespace
  - v0.2.0 - public preview release
  - v1.0.0 - BAD TAG on an initial version, fixed to 0.1.0
  - noversion, based on an old C/perl code.


### License

MIT. See LICENSE file.

(c)2019 Wojciech S. Czarnecki, OHIR-RIPE

