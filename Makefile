tag:
	@git tag `grep -P '^\tversion = ' gig.go|cut -f2 -d'"'`
	@git tag|grep -v ^v
