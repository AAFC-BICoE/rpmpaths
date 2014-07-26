rpmpaths
========

Scans a list of RPMs and generates PATH and :LD_LIBRARY_PATH based on locations of executables and libraries

Usage
======
```
	./rpmpaths: Find executables and libraries in RPMs to populate PATH and LD_LIBRARY_PATH
Usage: ./rpmpaths <rpmfile0> ... <rpmfileN>
	 Returns 2 lines with each of the following followed by the paths found for each:  "LD PATH: "   "PATH: " 
```


Copyright
========
Copyright (c) 2014 Government of Canada
Developed at Agriculture and Agri-Food Canada
Developed by Glen Newton
MIT License (Open Source)
