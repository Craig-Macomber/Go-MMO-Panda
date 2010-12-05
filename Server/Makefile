include $(GOROOT)/src/Make.inc

# app name
TARG=main

GOFILES=\
	main.go\

# dependancies
dirs:=tcp

# Copyright 2009 The Go Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

ifeq ($(GOOS),windows)
TARG:=$(TARG).exe
endif

all: SubPackages $(TARG)

include $(QUOTED_GOROOT)/src/Make.common

PREREQ+=$(patsubst %,%.make,$(DEPS))




# replace the target in $(GOROOT)/src/Make.cmd to add sub package dependancies
files:= $(foreach dir,$(dirs),$(dir)/_go_.$O)
lfolder:=$(foreach dir,$(dirs),$(dir)/_obj)
comma := ,
space :=
space +=
lfolder:=$(subst $(space),$(comma),$(lfolder))

$(TARG): _go_.$O $(lfolder)
	$(LD) -o $@ -L $(lfolder) _go_.$O 


SubPackages:
	$(foreach dir,$(dirs),gomake -C $(dir);) 
    


_go_.$O: $(GOFILES) $(PREREQ)  $(lfolder)
	$(GC) -o $@ -I $(lfolder) $(GOFILES)

# a recursive clean that cleans subpackages
rclean: clean
	$(foreach dir,$(dirs),gomake -C $(dir) rclean;) 


install: $(QUOTED_GOBIN)/$(TARG)

$(QUOTED_GOBIN)/$(TARG): $(TARG)
	cp -f $(TARG) $(QUOTED_GOBIN)

CLEANFILES+=$(TARG)

nuke: clean
	rm -f $(QUOTED_GOBIN)/$(TARG)

run:
	./$(TARG)