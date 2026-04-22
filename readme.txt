docker build --platform linux/arm64  -t bassista/xdccdl:test .
docker run --rm --name xb_test --network host -d --dns 8.8.8.8 -v /mnt/extSSD/downloads/xdccdl:/root bassista/xdccdl:test
dex
xdcc-browse --username nellodl --channel-join-delay 2 --wait-time 3 --quiet --throttle 50M --search-engine xdcc-eu "challangers ita"


Modificato il file setup.py aggiungendo in coda all'array install_requires "urllib3<2"
che diventa quindi:
        install_requires=[
            "bs4",
            "requests",
            "cfscrape",
            "typing",
            "colorama",
            "irc",
            "puffotter",
            "sentry-sdk",
            "names",
            "urllib3<2"
        ],
altrimenti il comando xdcc-search falliva

irc.rizon.net server di default

xdcc-dl "/msg TLT|DVD-AM|04 xdcc send #116" --server irc.openjoke.org
xdcc-dl --username nellodl --channel-join-delay 2 --wait-time 3 --quiet --throttle 50M "Baby.Reindeer.1x01.Episodio.1.ITA"

#comando per avviare il container se non è già running
docker start xdccdl || docker run --rm --name xdccdl -d --dns 8.8.8.8 -v /mnt/extSSD/downloads/xdccdl:/root docker.io/bassista/xdccdl:dev
#comando per avviare un download specifiando server, bot e pacchetto
docker exec -it xdccdl xdcc-dl "/msg $2 xdcc send #$3" --username nellodl --server $1 --channel-join-delay 2 --wait-time 3 --quiet --throttle 50M
#in alternativa a quiet si può usare silen, verbose, debug
#--fallback-channel se è necessario specificare il canale
#--timeout se si vuol specificare un timeout entro cui il download deve iniziare

#per cercare pacchetti usare xdcc-search; usage: xdcc-search [-h] [-q] [-v] [-d] [--silent] search_term {nibl,subsplease,xdcc-eu,ixirc}
xdcc-search REINDEER xdcc-eu

#per cercare e scaricare un risultato della ricerca usare xdcc-browse che ha stessa sintassi di xdcc-dl 
#tranne che invece di prendere in input il msg da mandare il bot prende una stringa da ricercare e 
#accetta il parametro --search-engine per specificare il motore di ricerca che deve usare
xdcc-browse --username nellodl --channel-join-delay 2 --wait-time 3 --quiet --throttle 50M --search-engine xdcc-eu "Baby Reindeer 01 ITA ARH"
