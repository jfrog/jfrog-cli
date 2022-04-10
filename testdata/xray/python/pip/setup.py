#!/usr/bin/env python

from setuptools import setup

setup(
    name='jfrog-python-audit',
    version='1.0',
    description='Project example for building Python project with JFrog products',
    author='JFrog',
    author_email='jfrog@jfrog.com',
    url='https://github.com/jfrog/project-examples',
    packages=[],
    install_requires=['PyYAML<=5.1', 'Werkzeug<=0.10', 'Jinja2<=2.8','urllib3 <=1.24','django <= 1.11.16'],
)
