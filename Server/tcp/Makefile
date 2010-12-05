include $(GOROOT)/src/Make.inc

TARG=tcp
GOFILES=\
	tcp.go\
	connectedBag.go

CLEANFILES+=connectedBag.go

include $(GOROOT)/src/Make.pkg

rclean: clean

connectedBag.go : ../iterbag.got
	gotgo -o=connectedBag.go -package-name=tcp ../iterbag.got Connected