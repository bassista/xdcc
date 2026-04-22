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
import sys
import logging
from logging.handlers import RotatingFileHandler
from typing import Callable, Optional, Union, List
from argparse import ArgumentParser, Namespace
from puffotter.os import makedirs


def cli_start(
        main_func: Union[
            Callable[[], None],
            Callable[[Namespace], None],
            Callable[[Namespace, logging.Logger], None]
        ],
        arg_parser: ArgumentParser,
        exit_msg: str = "Goodbye",
        package_name: Optional[str] = None,
        sentry_dsn: Optional[str] = None,
        release_name: Optional[str] = None
):
    """
    Starts a program and sets up logging, as well as sentry error tracking
    :param main_func: The main function to call
    :param arg_parser: The argument parser to use
    :param exit_msg: The message printed when the program's execution is
                     stopped using a keyboard interrupt
    :param package_name: The package name of the application
    :param sentry_dsn: The sentry DSN to use
    :param release_name: The name of the release
    :return: None
    """
    try:
        args = arg_parser.parse_args()
        setup_logging(args, package_name)

        if release_name is None:
            if package_name is not None:
                import pkg_resources
                version = pkg_resources.get_distribution(package_name).version
                release_name = package_name + "-" + version
            else:
                release_name = "Unknown"
                package_name = "unknown"

        if sentry_dsn is not None:
            import sentry_sdk
            sentry_sdk.init(sentry_dsn, release=release_name)

        from inspect import signature
        sign = signature(main_func)
        if len(sign.parameters) == 0:
            main_func()  # type: ignore
        elif len(sign.parameters) == 1:
            main_func(args)  # type: ignore
        elif len(sign.parameters) == 2:
            logger = logging.getLogger(package_name)
            main_func(args, logger)  # type: ignore
        else:
            print("Invalid amount of parameters for main function")
    except KeyboardInterrupt:
        print(exit_msg)


def argparse_add_verbosity(parser: ArgumentParser):
    """
    Adds --quiet, --verbose and --debug parameters to an ArgumentParser
    :param parser: the parser to which to add those flags
    :return: None
    """
    parser.add_argument("-q", "--quiet", action="store_true",
                        help="Sets the verbosity level of the program to "
                             "'quiet'")
    parser.add_argument("-v", "--verbose", action="store_true",
                        help="Sets the verbosity level of the program to "
                             "'verbose'")
    parser.add_argument("-d", "--debug", action="store_true",
                        help="Sets the verbosity level of the program to "
                             "'debug'")
    parser.add_argument("--silent", action="store_true",
                        help="Silences all output to STDOUT")


def argparse_add_logfile(parser: ArgumentParser):
    """
    Adds the --logfile argument to the argument parser
    :param parser: The argument parser to modify
    :return: None
    """
    parser.add_argument("--logfile",
                        help="Specifies the location of a logfile")


def setup_logging(args: Namespace, package_name: Optional[str]):
    """
    Sets up logging for the provided arguments
    :param args: The CLI arguments
    :param package_name: The package name
    :return: None
    """

    if "silent" in args and args.silent:
        loglevel = 100  # Disable all logging output
        sys.stdout = open(os.devnull, "w")  # Disable all print output
    elif "quiet" in args and args.quiet:
        loglevel = logging.ERROR
    elif "verbose" in args and args.verbose:
        loglevel = logging.INFO
    elif "debug" in args and args.debug:
        loglevel = logging.DEBUG
    else:
        loglevel = logging.WARNING

    file_formatter = logging.Formatter(
        "[%(asctime)s] %(levelname)s:[%(filename)s:%(funcName)s(%(lineno)d)] "
        "%(message)s",
        datefmt="%Y-%m-%d:%H-%M-%S")

    stream_formatter = logging.Formatter("%(levelname)s: %(message)s")
    stream_handler = logging.StreamHandler()
    stream_handler.setFormatter(stream_formatter)
    stream_handler.setLevel(loglevel)

    handlers = [stream_handler]  # type: List[logging.Handler]

    if "logfile" in args and args.logfile is not None:
        log_file = args.logfile
        file_handler = logging.FileHandler(log_file)
        file_handler.setFormatter(file_formatter)
        file_handler.setLevel(logging.DEBUG)
        handlers.append(file_handler)

    if package_name is not None:
        log_dir = os.path.join(
            os.path.expanduser("~"), ".config", package_name
        )
        makedirs(log_dir)
        log_file = os.path.join(log_dir, "logs.log")
        rotating_handler = RotatingFileHandler(
            log_file, maxBytes=50000000, backupCount=10
        )
        rotating_handler.setFormatter(file_formatter)
        rotating_handler.setLevel(logging.DEBUG)
        handlers.append(rotating_handler)

    logging.basicConfig(level=logging.DEBUG, handlers=handlers)
