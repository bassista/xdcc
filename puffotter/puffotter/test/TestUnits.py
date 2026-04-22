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
from unittest import TestCase
from puffotter.units import human_readable_bytes, byte_string_to_byte_count


class TestUnits(TestCase):
    """
    Tests functions that handle units conversion
    """

    def test_byte_conversion(self):
        """
        Tests that byte strings can be parsed correctly and displayed in
        a human-readable format
        :return: None
        """
        for string, count in [
            ("1MB", 1000000),
            ("1.024KB", 1024),
            ("0.123KB", 123),
            ("1.234GB", 1234000000)
        ]:
            self.assertEqual(string, human_readable_bytes(count))
            self.assertEqual(byte_string_to_byte_count(string), count)

        # Test if human-readable strings are rounded correctly
        self.assertEqual("1.234GB", human_readable_bytes(1234123123))
