// Copyright (c) 2017 Townsourced Inc.

var gulp = require('gulp');
var path = require('path');
var sass = require('gulp-sass');
var sourcemaps = require('gulp-sourcemaps');

var deployDir = './deploy';
var dev = false;

// gulp.task('js', function (callback) {
//     // rollup
//     // buble
//     // copy vue dist to static
// });


gulp.task('css', function () {
    return gulp.src('./scss/**/*.scss')
        .pipe(sass({
            outputStyle: 'compressed',
            includePaths: 'node_modules'
        }).on('error', sass.logError))
        .pipe(gulp.dest(path.join(deployDir, 'css')));
});

gulp.task('devCss', function () {
    return gulp.src('./scss/**/*.scss')
        .pipe(sourcemaps.init())
        .pipe(sass({
            outputStyle: 'compressed',
            includePaths: 'node_modules'
        }).on('error', sass.logError))
        .pipe(sourcemaps.write())
        .pipe(gulp.dest(path.join(deployDir, 'css')));
});


// static files
gulp.task('html', function () {
    return gulp.src([
        './**/*.html',
        '!deploy/**/*',
        '!node_modules/**/*'
    ]).pipe(gulp.dest(deployDir));
});

gulp.task('images', function () {
    return gulp.src('./images/*')
        .pipe(gulp.dest(path.join(deployDir, 'images')))
});


// watch for changes
gulp.task('watch', function () {
    gulp.watch('./**/*.html', ['html']);
    gulp.watch('./images/**/*', ['images']);
    gulp.watch(['./scss/**/*.scss'], ['devCss']);
    // gulp.watch(['./src/ts/**/*.ts', './src/ts/**/*.vue'], ['js']);
});


gulp.task('dev', ['html', 'images', 'devCss']);

// start default task
gulp.task('default', ['html', 'images', 'css']);