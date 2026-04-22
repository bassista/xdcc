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

import logging
from colorama import Back, Fore, Style
from typing import Optional


class ColorLogger:

    def __init__(
            self,
            logger: logging.Logger,
            debug_bg: str = Back.WHITE,
            debug_fg: str = Fore.BLACK,
            info_bg: str = Back.BLACK,
            info_fg: str = Fore.WHITE,
            warning_bg: str = Back.YELLOW,
            warning_fg: str = Fore.BLACK,
            error_bg: str = Back.LIGHTRED_EX,
            error_fg: str = Fore.BLACK
    ):
        self.logger = logger
        self.debug_style = debug_bg + debug_fg
        self.info_style = info_bg + info_fg
        self.warning_style = warning_bg + warning_fg
        self.error_style = error_bg + error_fg

    def __format_message(
            self,
            level: int,
            message: str,
            bg: Optional[str],
            fg: Optional[str]
    ):
        """
        Formats a message
        :param level: The message's level
        :param message: The message to format
        :param bg: Optional background style override
        :param fg: Optional foreground style override
        :return: The formatted message
        """
        style = ""
        if bg is not None:
            style += bg
        if fg is not None:
            style += fg
        if style == "":
            style = {
                logging.DEBUG: self.debug_style,
                logging.INFO: self.info_style,
                logging.WARNING: self.warning_style,
                logging.ERROR: self.error_style
            }[level]
        return style + message + Style.RESET_ALL

    def debug(
            self,
            message: str,
            bg: Optional[str] = None,
            fg: Optional[str] = None
    ):
        """
        Logs a message at DEBUG level
        :param message: The message to log
        :param bg: Overrides the default background style
        :param fg: Overrides the default foreground style
        :return: None
        """
        self.logger.debug(
            self.__format_message(logging.DEBUG, message, bg, fg)
        )

    def info(
            self,
            message: str,
            bg: Optional[str] = None,
            fg: Optional[str] = None
    ):
        """
        Logs a message at INFO level
        :param message: The message to log
        :param bg: Overrides the default background style
        :param fg: Overrides the default foreground style
        :return: None
        """
        self.logger.info(self.__format_message(logging.INFO, message, bg, fg))

    def warning(
            self,
            message: str,
            bg: Optional[str] = None,
            fg: Optional[str] = None
    ):
        """
        Logs a message at WARNING level
        :param message: The message to log
        :param bg: Overrides the default background style
        :param fg: Overrides the default foreground style
        :return: None
        """
        self.logger.warning(
            self.__format_message(logging.WARNING, message, bg, fg)
        )

    def error(
            self,
            message: str,
            bg: Optional[str] = None,
            fg: Optional[str] = None
    ):
        """
        Logs a message at ERROR level
        :param message: The message to log
        :param bg: Overrides the default background style
        :param fg: Overrides the default foreground style
        :return: None
        """
        self.logger.error(
            self.__format_message(logging.ERROR, message, bg, fg)
        )
