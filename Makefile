clean:
	# clean tmp folders
	rm -rf tmp/node_{a..d}

prepare: clean
	mkdir -p tmp/node_{a..d}

