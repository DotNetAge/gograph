/*
 * GoGraph SWIG Interface File
 * 
 * This file defines how SWIG generates bindings for different languages.
 * Modify this file to customize the generated wrappers.
 */

%module(directors="1") gograph

/* Enable CGO support */
%{
#include "gograph_c.h"
%}

/* Include the C header file */
%include "gograph_c.h"

/* Configure Go package */
%goheader("package binding")
%gopackage("binding")

/* ============================================================================
 * Type Maps - Customize type conversions
 * ============================================================================ */

/* Handle const char* strings */
%typemap(in) const char* {
    if ($input == NULL) {
        PyErr_SetString(PyExc_TypeError, "string expected");
        return NULL;
    }
    $1 = $input;
}

/* Handle output parameters */
%typemap(out) uint64_t* {
    PyObject* result = Py_BuildValue("(K)", (unsigned long long)*$1);
    $result = result;
}

/* Handle ErrorInfo* as output parameter */
%typemap(in, numinputs=0) ErrorInfo* OUTPUT(ErrorInfo arg) {
    $1 = &arg;
}

/* Handle QueryResult* as output parameter */
%typemap(in, numinputs=0) QueryResult* OUTPUT(QueryResult arg) {
    $1 = &arg;
}

/* Handle Node* as output parameter */
%typemap(in, numinputs=0) Node* OUTPUT(Node arg) {
    $1 = &arg;
}

/* Handle Relationship* as output parameter */
%typemap(in, numinputs=0) Relationship* OUTPUT(Relationship arg) {
    $1 = &arg;
}
