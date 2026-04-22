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


def load_env_file(path: str = ".env"):
    """
    Loads an environment file into os.environ
    :param path: The path to the environment file
    :return: None
    """
    if os.path.isfile(path):
        with open(path, "r") as f:
            for line in f.read().split("\n"):
                try:
                    key, value = line.split("#")[0].split("=", 1)
                    os.environ[key] = value
                except ValueError:
                    pass
