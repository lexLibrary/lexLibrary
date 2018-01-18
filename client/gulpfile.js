// Copyright (c) 2017 Townsourced Inc.

var gulp = require('gulp');
var path = require('path');
var sass = require('gulp-sass');
var postcss = require('gulp-postcss');
var autoprefixer = require('autoprefixer');
var sourcemaps = require('gulp-sourcemaps');
var rename = require('gulp-rename');
var rollup = require('gulp-rollup');
var uglify = require('rollup-plugin-uglify');
var buble = require('rollup-plugin-buble');

var deployDir = './deploy';

var rollupCFG = {
    input: [
        'js/login.js',
    ],
    format: 'iife',
    plugins: [
        uglify(),
        buble(),
    ]
};

gulp.task('js', function (callback) {
    //     // buble
    return [
        gulp.src('./js/**/*.js')
            .pipe(rollup(rollupCFG))
            .pipe(gulp.dest(path.join(deployDir, 'js'))),
        gulp.src('./node_modules/vue/dist/vue.min.js')
            .pipe(rename('vue.js'))
            .pipe(gulp.dest(path.join(deployDir, 'js')))
    ];
});

gulp.task('devJs', function (callback) {
    return [
        gulp.src('./js/**/*.js')
            .pipe(sourcemaps.init())
            .pipe(rollup(rollupCFG))
            .pipe(sourcemaps.write())
            .pipe(gulp.dest(path.join(deployDir, 'js'))),
        gulp.src('./node_modules/vue/dist/vue.js')
            .pipe(gulp.dest(path.join(deployDir, 'js')))
    ];
});

gulp.task('css', function () {
    return gulp.src('./scss/**/*.scss')
        .pipe(sass({
            outputStyle: 'compressed',
            includePaths: 'node_modules'
        }).on('error', sass.logError))
        .pipe(postcss([
            autoprefixer({
                cascade: false
            })
        ]
        ))
        .pipe(gulp.dest(path.join(deployDir, 'css')));
});

gulp.task('devCss', function () {
    return gulp.src('./scss/**/*.scss')
        .pipe(sourcemaps.init())
        .pipe(sass({
            outputStyle: 'compressed',
            includePaths: 'node_modules'
        }).on('error', sass.logError))
        .pipe(postcss([
            autoprefixer({
                cascade: false
            })
        ]
        ))
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
    gulp.watch(['./js/**/*.js'], ['devJs']);
});

gulp.task('dev', ['html', 'images', 'devCss', 'devJs']);

// start default task
gulp.task('default', ['html', 'images', 'css', 'js']);
