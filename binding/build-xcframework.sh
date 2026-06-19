#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUTPUT_DIR="${PROJECT_ROOT}/dist"
BUILD_DIR="${PROJECT_ROOT}/binding/build"
XCFRAMEWORK_NAME="CGograph"
XCFRAMEWORK_PATH="${OUTPUT_DIR}/${XCFRAMEWORK_NAME}.xcframework"
ARCHS=("arm64" "amd64")

echo "==> Cleaning..."
rm -rf "${BUILD_DIR}" "${XCFRAMEWORK_PATH}"
mkdir -p "${BUILD_DIR}" "${OUTPUT_DIR}"

# Build Go as c-archive for each architecture
for arch in "${ARCHS[@]}"; do
    echo "==> Building for ${arch}..."
    (
        cd "${PROJECT_ROOT}"
        GOOS=darwin GOARCH="${arch}" CGO_ENABLED=1 \
            go build -buildmode=c-archive \
            -o "${BUILD_DIR}/libgograph_${arch}.a" \
            ./binding/cmd
    )
    echo "    -> ${BUILD_DIR}/libgograph_${arch}.a"
done

# Create universal binary with lipo
echo "==> Creating universal binary..."
lipo -create \
    "${BUILD_DIR}/libgograph_arm64.a" \
    "${BUILD_DIR}/libgograph_amd64.a" \
    -output "${BUILD_DIR}/libgograph.a"
echo "    -> ${BUILD_DIR}/libgograph.a"

# Copy header
cp "${PROJECT_ROOT}/binding/gograph_c.h" "${BUILD_DIR}/"

# Create xcframework
echo "==> Creating xcframework..."
xcodebuild -create-xcframework \
    -library "${BUILD_DIR}/libgograph.a" \
    -headers "${BUILD_DIR}/gograph_c.h" \
    -output "${XCFRAMEWORK_PATH}"
echo "    -> ${XCFRAMEWORK_PATH}"

echo ""
echo "Done! You can now add ${XCFRAMEWORK_PATH} to your Xcode project."
echo "Alternatively, use the SPM package at: ${OUTPUT_DIR}/Package.swift"
