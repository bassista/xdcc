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

import os
from typing import List
from subprocess import check_call, CalledProcessError, STDOUT
from puffotter.print import pprint


def execute_command(command: List[str]) -> int:
    """
    Executes a command
    :param command: The command to execute
    :return: The status code
    """
    pprint(" ".join(command), fg="lyellow")
    with open(os.devnull, "w") as devnull:
        try:
            code = check_call(command, stdout=devnull, stderr=STDOUT)
            if code != 0:
                pprint("Error Code {}".format(code), fg="lred")
            return code
        except CalledProcessError as e:
            pprint("Called Process Error: " + str(e), fg="lred")
            return 1
