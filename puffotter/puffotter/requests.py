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

import time
import requests
import logging


def aggressive_request(url: str) -> str:
    """
    Handles GET requests while analyzing status codes
    :param url: The URL to get
    :return: The response text
    """
    logger = logging.getLogger("puffotter")

    time.sleep(1)
    headers = {"User-Agent": "Mozilla/5.0"}
    resp = requests.get(url, headers=headers)

    while resp.status_code >= 300:
        logger.warning("HTTP Error Code: {}".format(resp.status_code))
        logger.debug(resp.headers)
        logger.debug(resp.text)
        time.sleep(60)
        resp = requests.get(url, headers=headers)

    return resp.text
