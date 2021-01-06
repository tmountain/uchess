TXTMAN=txt2man

doc: docs/uchess.txt
	$(TXTMAN) -s1 -p -P uchess -t uchess docs/uchess.txt > docs/uchess.man
