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

import requests


def verify_recaptcha(
        client_ip: str,
        recaptcha_response: str,
        secret_key: str
) -> bool:
    """
    Verifies a recaptcha response.
    If the recaptcha response originates from a local address,
    this method will always return True.
    :param client_ip: The IP Address of the client solving the captcha
    :param recaptcha_response: the recaptcha response to verify
    :param secret_key: the recaptcha secret key
    :return: True if the recaptcha response was correct, False otherwise
    """
    return requests.post(
        "https://www.google.com/recaptcha/api/siteverify",
        data={
            "secret": secret_key,
            "response": recaptcha_response,
            "remoteip": client_ip
        }
    ).json().get("success", False)
