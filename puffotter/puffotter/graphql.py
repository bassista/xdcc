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

import json
import requests
from typing import Dict, Any, Optional


class GraphQlClient:
    """
    A simple API wrapper for GraphQL APIs
    """

    def __init__(self, api_url: str):
        """
        Initializes the GraphQL API wrapper
        :param api_url: The API endpoint URL
        """
        self.api_url = api_url

    def query(
            self,
            query_string: str,
            variables: Optional[Dict[str, Any]]
    ) -> Optional[Dict[str, Any]]:
        """
        Executes a GraphQL query
        :param query_string: The query string to use
        :param variables: The variables to send
        :return: The response JSON, or None if an error occurred.
        """
        if variables is None:
            variables = {}

        resp = requests.post(self.api_url, json={
            "query": query_string,
            "variables": variables
        })
        if not resp.status_code < 300:
            return None
        else:
            return json.loads(resp.text)
