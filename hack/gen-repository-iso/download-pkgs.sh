#! /bin/sh

for p in ${PACKAGES} ; do
  echo "\n Download $p ... \n"
  sudo apt-get download $p 2>>errors.txt
  for i in $(apt-cache depends $p | grep -E 'Depends|Recommends|Suggests' | cut -d ':' -f 2,3 | sed -e s/' '/''/); do sudo apt-get download $i 2>>errors.txt; done
done
