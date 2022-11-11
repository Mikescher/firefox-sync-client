Virus Scanner false positives under Windows
===========================================


The windows binaries get flagged by some virus scanners (especially the 32 bit exe).  
These are false positives and it seems I can't remove them without signing the binaries.

Apparently this is a known issue, and some virus scanner heuristics are especially trigger happy with golang binaries because some high-profile malware was written in it...

If you are worried I recommend to look at the source code in this repository and then build it yourself (see the Makefile)
