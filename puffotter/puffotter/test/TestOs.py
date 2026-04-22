"""LICENSE
Copyright 2017 Hermann Krumrey <hermann@krumreyh.com>

This file is part of bundesliga-tippspiel.

bundesliga-tippspiel is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

bundesliga-tippspiel is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with bundesliga-tippspiel.  If not, see <http://www.gnu.org/licenses/>.
LICENSE"""

# noinspection PyProtectedMember
import os
import shutil
from unittest import TestCase
from puffotter.os import listdir


class TestCrypto(TestCase):
    """
    Tests os functions
    """

    def test_listdir(self):
        """
        Tests the listdir function
        :return: None
        """
        os.makedirs("test")
        os.makedirs("test/1")
        os.makedirs("test/.2")
        open("test/3", "w").close()
        open("test/.4", "w").close()

        _all = listdir("test", no_dot=False)
        self.assertEqual(len(_all), 4)

        no_dot = listdir("test")
        self.assertEqual(len(no_dot), 2)

        no_files = listdir("test", no_files=True)
        self.assertEqual(len(no_files), 1)
        self.assertEqual(no_files[0], ("1", "test/1"))

        no_dirs = listdir("test", no_dirs=True)
        self.assertEqual(len(no_dirs), 1)
        self.assertEqual(no_dirs[0], ("3", "test/3"))

        nothing = listdir("test", no_files=True, no_dirs=True)
        self.assertEqual(len(nothing), 0)

        shutil.rmtree("test")
