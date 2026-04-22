"""LICENSE
Copyright 2016 Hermann Krumrey <hermann@krumreyh.com>

This file is part of xdcc-dl.

xdcc-dl is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

xdcc-dl is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with xdcc-dl.  If not, see <http://www.gnu.org/licenses/>.
LICENSE"""

import requests
from typing import List
from bs4 import BeautifulSoup
from xdcc_dl.entities.XDCCPack import XDCCPack
from xdcc_dl.entities.IrcServer import IrcServer
from puffotter.units import byte_string_to_byte_count

def remove_non_numeric_and_keep_digits(input_string):
    result = ""
    found_digit = False

    for char in input_string:
        # Se trovi un carattere numerico, cambia il flag e aggiungi il numero alla stringa risultante
        if char.isdigit():
            found_digit = True
            result += char
        # Se il flag è già impostato su True, aggiungi il carattere alla stringa risultante anche se non è un numerico
        elif found_digit:
            result += char
        # Ignora i caratteri non numerici prima di trovare il primo carattere numerico
        else:
            continue
    return result

def get_value_for_param(param_name, input_string):
    for token in input_string:
         if token.startswith(param_name):
            token_split = token.split('=')
            if len(token_split) >= 2:
               return token_split[1]

    return ""


def checkType(type_search, filename):
    if type_search == "v":
       if filename.lower().endswith("mkv") or filename.lower().endswith("avi") or filename.lower().endswith("mp4"):
          return True
       else:
          return False
    return True


def checkBot(bot_name, bot):
    if bot_name:
       if bot.lower().__contains__(bot_name.lower()):
          return True
       else: return False
    return True


def checkSize(size_type, size):
    return True

def find_xdcc_eu_packs(search_phrase: str) -> List[XDCCPack]:
    # print(search_phrase)
    """
    Se nella stringa di ricerca è presente il carattere pipe '|'
    allora oltre alla query di ricerca sono presenti anche uno o più filtri di ricerca
    con 't' viene indicata l'estensione del filename che si cerca
    con 'b' viene indicato il nome del bot da cui si vuole il file (non il nome completo ma una sottostringa tipo TLT)
    con 's' viene indicato il tipo di dimensione file che si cerca (i.e. Megabyte o Gigabyte)
    con 'q' viene indicata la query di ricerca vera e propria
    """

    type_search=""
    bot_name=""
    size_type=""
    if search_phrase.__contains__('|'):
       tokens = search_phrase.split('|')
       type_search = get_value_for_param('t', tokens)
      # print(type_search)
       search_phrase = get_value_for_param('q', tokens)
      # print(search_phrase)
       bot_name = get_value_for_param('b', tokens)
      # print(bot_name)
       size_type = get_value_for_param('s', tokens)
      # print(size_type)

    """
    Method that conducts the xdcc pack search for xdcc.eu

    :return: the search results as a list of XDCCPack objects
    """
    url = "https://www.xdcc.eu/search.php?searchkey=" + search_phrase
    response = requests.get(url)
    soup = BeautifulSoup(response.text, "html.parser")
    entries = soup.select("tr")

    try:
       entries.pop(0)
    except Exception:
        print("Nothing found, check your search parameters")

    packs = []
    for entry in entries:
        parts = entry.select("td")
        info = parts[1].select("a")[1]
        server = IrcServer(info["data-s"])
        pack_message = info["data-p"]
        bot, pack_number = pack_message.split(" xdcc send #")

        #print(parts[5].text)
        size = byte_string_to_byte_count(remove_non_numeric_and_keep_digits(parts[5].text))
        filename = parts[6].text

        pack = XDCCPack(server, bot, int(pack_number))
        pack.set_size(size)
        pack.set_filename(filename)

        if not type_search and not bot_name and not size_type:
           packs.append(pack)
        elif checkType(type_search, filename) and checkBot(bot_name, bot) and checkSize(size_type, size):
           packs.append(pack)

    return packs
