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


### Sample config in OCONF format:

```
// line comment can also begin with " # ! (d-quote, sharp and bang) markers.
   free comment lines are possible too - these may not contain ' : ' though.

 Section : >                        //  >  is a "section" marker and depth indicator. 
    spaced :  val & spaces     |.   //  |  "guard pragma" keeps tail blanks intact.
    noComm : hello // there    '.   //  '  "disa  pragma" makes former // to the Value
   withCTL : Use\t tab and \n  \.   //  \  "unesc pragma" unescapes \t and \n
    withNL : some value        ^.   //  ^  "newline pragma" adds a '\n' at the end.
    looong : value can span    +.   //  +  "join Pragma" joins this line value with
           :  many lines and   +.   //      next line value 
           :: still keep indent.    // ::  separator makes leading space more visible.

  SubSec : >>                       // >>  open subsection at depth 2.
            : list member  0        //     Ordered (unnamed) values can be indexed 
            : list member  1        //     naturally by the order of apperance
        33  : list member 33        //     or with index being given explicit
            : value                 //     /Section/SubSect[34] = value

   SSSub : >>>                      // >>>  go depth 3 sub-sub section
         key : value                //      /Section/SubSect/SSSub/key = value

 OthSect : >                        // Next depth 1 section opens. All above close.
       key : value                  //    /OthSect/key = value
     a key : value                  // spaces in keys are ok.   Here is  "a key".
   ' spkey :  value                 // '  quotes leading space. Here is " spkey".
       Имя : Юрий                   // OConf supports utf-8 encoded unicode
       键k : v值                     // in full range. [And 8bit "codepages" too].

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

  - v1.0.0 - first public release


### License

MIT. See LICENSE file.

(c)2019 Wojciech S. Czarnecki, OHIR-RIPE

