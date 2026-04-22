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

import imaplib


def get_inbox_count(
        imap_server: str,
        imap_address: str,
        imap_password: str,
        imap_port: int = 993
) -> int:
    """
    Checks the amount of emails in an IMAP inbox
    :param imap_server: The IMAP server to use
    :param imap_address: The IMAP address to use
    :param imap_password: The IMAP password to use
    :param imap_port: The IMAP port to use
    :return: The amount of emails
    """
    server = imaplib.IMAP4_SSL(imap_server, imap_port)
    server.login(imap_address, imap_password)
    content = server.select("Inbox")[1][0]
    counted = 0 if content is None else int(content)
    server.close()
    server.logout()
    return counted
