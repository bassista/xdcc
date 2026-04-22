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

import bcrypt
import random
import string


def generate_random(length: int) -> str:
    """
    Generates a random byte string consisting of alphanumeric characters
    Thanks @ Albert (https://stackoverflow.com/users/281021/albert)
    https://stackoverflow.com/questions/2257441
    :param length: The length of the string to generate
    :return: The generated random byte string
    """
    return bytes(
        "".join(random.choice(string.ascii_uppercase + string.digits)
                for _ in range(length)),
        "utf-8"
    ).decode("utf-8")


def generate_hash(password: str) -> str:
    """
    Salts and hashes a password to generate a hash for storage in a database
    :param password: The password to hash
    :return: The hash of the password
    """
    password_bytes = bytes(password, "utf-8")
    return bcrypt.hashpw(password_bytes, bcrypt.gensalt()).decode("utf-8")


def verify_password(password: str, hashed: str):
    """
    Verifies that a password matches a given hash
    :param password: The password to verify
    :param hashed: The hash to verify the password against
    :return: True if the password matches, otherwise False
    """
    password_bytes = bytes(password, "utf-8")
    hashed_bytes = bytes(hashed, "utf-8")

    try:
        return bcrypt.checkpw(password_bytes, hashed_bytes)
    except ValueError:
        return False
