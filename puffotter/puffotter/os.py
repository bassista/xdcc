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
import shutil
from pathlib import Path
from typing import Tuple, List, Optional


def listdir(
        path: str,
        no_files: bool = False,
        no_dirs: bool = False,
        no_dot: bool = True
) -> List[Tuple[str, str]]:
    """
    Improves on the standard os.listdir function.
    By default, files and directories starting with `.` are ignored and
    one can disable checking for files or directories.
    Instead of just the filenames, the function returns tuples of filenames
    and relative file paths.
    :param path: The path of which to list the contents
    :param no_files: If set to True, will ignore files
    :param no_dirs: If set to True, will ignore directories
    :param no_dot: If set to True, will ignore files starting with `.`
    :return: A sorted list of the contents of the directory, consisting of
             tuples of the file/directory name and their relative path.
    """
    content = []
    for child in sorted(os.listdir(path)):
        child_path = os.path.join(path, child)

        if no_files and os.path.isfile(child_path):
            continue
        elif no_dirs and os.path.isdir(child_path):
            continue
        elif no_dot and child.startswith("."):
            continue
        else:
            content.append((child, child_path))
    return content


def makedirs(path: str, delete_before: bool = False):
    """
    A more pleasant to use makedirs function.
    Only calls os.makedirs if the directory does not already exist
    :param path: The path to the directory to create
    :param delete_before: If True, deletes the directory beforehand,
                          if it exists
    :return: None
    """
    if delete_before and os.path.isdir(path):
        shutil.rmtree(path)

    if not os.path.isdir(path):
        os.makedirs(path)


def replace_illegal_ntfs_chars(string: str) -> str:
    """
    Replaces illegal characters in NTFS file systems with
    appropriate substitutes
    :param string: Te string in which to replace the illegal characters
    :return: The sanitized string
    """
    illegal_characters = {
        "/": "ǁ",
        "\\": "ǁ",
        "?": "‽",
        "<": "←",
        ">": "→",
        ":": ";",
        "*": "∗",
        "|": "ǁ",
        "\"": "“"
    }
    for illegal_character, replacement in illegal_characters.items():
        string = string.replace(illegal_character, replacement)
    return string


def create_file(path: str):
    """
    Creates an empty file
    :param path: The path to the file
    :return: None
    """
    if not os.path.isfile(path):
        with open(path, "w") as f:
            f.write("")


def get_ext(filename: str) -> Optional[str]:
    """
    Gets the file extension of a file
    :param filename: The filename for which to get the file extension
    :return: The file extension or None if the file has no extension
    """
    try:
        return filename.rsplit(".", 1)[1]
    except IndexError:
        return None


def touch(path: str):
    """
    Ensures that a file exists
    :param path: Path to the file to ensure it exists
    """
    Path(path).touch()
