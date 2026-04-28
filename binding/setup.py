"""
GoGraph Python Binding Setup

This setup.py builds a Python package for GoGraph, a high-performance graph database.
"""

import os
import sys
from setuptools import setup, find_packages
from setuptools.extension import Extension
from setuptools.command.build_ext import build_ext

__version__ = "0.2.2"

class CustomBuildExt(build_ext):
    """Custom build extension to copy pre-built shared library"""
    
    def run(self):
        # Copy pre-built _gograph.so to the build directory
        import shutil
        
        src_path = os.path.join(os.path.dirname(__file__), '_gograph.so')
        if os.path.exists(src_path):
            for ext in self.extensions:
                target_path = self.get_ext_fullpath(ext.name)
                target_dir = os.path.dirname(target_path)
                os.makedirs(target_dir, exist_ok=True)
                shutil.copy(src_path, target_path)
                print(f"Copied pre-built {src_path} to {target_path}")

setup(
    name="gograph",
    version=__version__,
    description="GoGraph is a lightweight, zero-dependency, embedded graph database written entirely in Go. Think of it as SQLite for Graph Databases",
    long_description=open("README.md", "r").read(),
    long_description_content_type="text/markdown",
    author="Ray",
    author_email="ray@rayainfo.cn",
    url="https://github.com/DotNetAge/gograph",
    packages=find_packages(),
    ext_modules=[
        Extension(
            "gograph._gograph",
            sources=[],  # Pre-built shared library
        )
    ],
    cmdclass={
        'build_ext': CustomBuildExt,
    },
    include_package_data=True,
    zip_safe=False,
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: MacOS",
        "Operating System :: POSIX",
        "Programming Language :: Python",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Programming Language :: Python :: 3.13",
        "Programming Language :: Python :: 3.14",
        "Topic :: Database",
        "Topic :: Database :: Database Engines/Servers",
    ],
    python_requires=">=3.9",
    install_requires=[],
)
