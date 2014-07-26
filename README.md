rpmpaths
========

Scans a list of RPMs and generates PATH and :LD_LIBRARY_PATH based on locations of executables and libraries

Usage
======
```
	./rpmpaths: Find executables and libraries in RPMs to populate PATH and LD_LIBRARY_PATH
	Usage: ./rpmpaths <-c> | <rpmfile0> ... <rpmfileN>
	Default: read the list of args as rpm files
	-c : Read each line from standard input as a list of rpm files

	Returns 2 lines with each of the following followed by the paths found for each:  "LD PATH: "   "PATH: " 

	 Example:
	 $ find . -name \*.rpm -print | rpmpaths -c
	 PATH=/opt/bio/gdal/bin:/opt/bio/geos/bin
	 LD_LIBRARY_PATH=/opt/bio/gdal/lib:/opt/bio/geos/lib:/opt/bio/zlib/lib:/usr/lib64

```

Copyright, License, Attribution& Acknowledgements
=====
* Copyright 2014 Government of Canada
* MIT License (See LICENSE file)
* Author: Glen Newton glen.newton@agr.gc.ca glen.newton@gmail.com
* Developed at: Microbial Biodiversity Bioinformatics Group @ Agriculture and Agri-Food Canada
