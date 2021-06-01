# sdutil

stdutil is a simple command line utility for interacting with stardict
dictionaries.

## List dictionaries

```
$ sdutil list /path/to/dictionaries
Name         4JWORDS
Author:      Kanji Haitani
Email:       khaitani1@earthlink.net
Word Count:  3183

Name         jmdict-en-ja
Author:      Jim Breen
Email:       j.breen@csse.monash.edu.au
Word Count:  95443

Name         jmdict-ja-en
Author:      Jim Breen
Email:       j.breen@csse.monash.edu.au
Word Count:  171879
```

## Search dictionaries

```
$ sdutil query /path/to/dictionary "dictionary"
jmdict-en-ja

(character) dictionary
字書
じしょ

a Muromachi era Japanese dictionary
節用集
せつようしゅう

based on (according to) the dictionary
辞書に拠れば
じしょによれば
...
```
