"""LICENSE
Copyright 2019 Hermann Krumrey <hermann@krumreyh.com>

This file is part of puffotter.

puffotter is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

puffotter is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with puffotter.  If not, see <http://www.gnu.org/licenses/>.
LICENSE"""

import sys
from typing import IO, Optional
from colorama import Style, Fore, Back


def pprint(
        *objects: str,
        sep: str = " ",
        end: str = "\n",
        file: IO = sys.stdout,
        flush: bool = False,
        fg: Optional[str] = None,
        bg: Optional[str] = None
):
    """
    Function that extends the print function
    :return: None
    """
    code = ""
    if fg is not None:
        code += __convert_color_to_code(fg, "fore")
    if bg is not None:
        code += __convert_color_to_code(bg, "back")

    if code != "":
        object_list = []
        for i in range(0, len(objects)):
            object_list.append(code + str(objects[i]) + Style.RESET_ALL)
    else:
        object_list = list(objects)

    print(*object_list, sep=sep, end=end, file=file, flush=flush)


def __convert_color_to_code(color_string: str, mode: str) -> str:
    """
    Converts a color string into the correct ANSI code
    If the provided string does not match one of the predefined color strings,
    the strng will be returned as is.
    :param color_string: The string to convert
    :param mode: The mode, either 'fore' or 'back'.
                 Otherwise a ValueError will be raised.
    :return: The ANSI code
    """
    color_chart = {
        "lred": "LIGHTRED_EX",
        "red": "RED",
        "lyellow": "LIGHTYELLOW_EX",
        "yellow": "YELLOW",
        "lblue": "LIGHTBLUE_EX",
        "blue": "BLUE",
        "lgreen": "LIGHTGREEN_EX",
        "green": "GREEN",
        "lmagenta": "LIGHTMAGENTA_EX",
        "magenta": "MAGENTA",
        "lcyan": "LIGHTCYAN_EX",
        "cyan": "CYAN",
        "lwhite": "LIGHTWHITE_EX",
        "white": "WHITE",
        "lblack": "LIGHTBLACK_EX",
        "black": "BLACK"
    }
    code = color_chart.get(color_string.lower())
    if code is None:
        return color_string
    elif mode == "fore":
        return getattr(Fore, code)
    elif mode == "back":
        return getattr(Back, code)
    else:
        raise ValueError("Invalid Mode")
