// Copyright (c) 2017-2018 Townsourced Inc.

const gulp = require('gulp');
const path = require('path');
const sass = require('gulp-sass');
const postcss = require('gulp-postcss');
const autoprefixer = require('autoprefixer');
const sourcemaps = require('gulp-sourcemaps');
const rename = require('gulp-rename');
const rollup = require('rollup');
const uglify = require('rollup-plugin-uglify');
const buble = require('rollup-plugin-buble');
const resolve = require('rollup-plugin-node-resolve');
const commonjs = require('rollup-plugin-commonjs');

var deployDir = './deploy';

var jsFiles = [
    'login.js',
    'index.js',
    'signup.js',
    'about.js',
    'first_run.js',
    'profile.js',
    'profile_edit.js',
    'admin.js',
];

function rollupFiles(dev) {
    let promises = [];
    for (let i = 0; i < jsFiles.length; i++) {
        promises.push(rollupFile('./js/' + jsFiles[i], dev));
    }
    return Promise.all(promises);
}

function rollupFile(file, dev) {
    let sourcemaps = false;
    let plugins = [
        resolve(),
        commonjs({}),
        buble(),
    ];
    if (dev) {
        sourcemaps = 'inline';
    } else {
        plugins.push(uglify());
    }
    return rollup.rollup({
            input: file,
            plugins: plugins
        })
        .then(bundle => {
            return bundle.write({
                file: path.join(deployDir, file),
                format: 'iife',
                sourcemap: sourcemaps,
            });
        });
}

gulp.task('js', function() {
    return [
        rollupFiles(),
        gulp.src('./node_modules/vue/dist/vue.min.js')
        .pipe(rename('vue.js'))
        .pipe(gulp.dest(path.join(deployDir, 'js')))
    ];
});

gulp.task('devJs', function() {
    return [
        rollupFiles(true),
        gulp.src('./node_modules/vue/dist/vue.js')
        .pipe(gulp.dest(path.join(deployDir, 'js')))
    ];
});

gulp.task('css', function() {
    return gulp.src('./scss/**/*.scss')
        .pipe(sass({
            outputStyle: 'compressed',
            includePaths: 'node_modules'
        }).on('error', sass.logError))
        .pipe(postcss([
            autoprefixer({
                cascade: false
            })
        ]))
        .pipe(gulp.dest(path.join(deployDir, 'css')));
});

gulp.task('devCss', function() {
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
        ]))
        .pipe(sourcemaps.write())
        .pipe(gulp.dest(path.join(deployDir, 'css')));
});


// static files
gulp.task('html', function() {
    return gulp.src([
        './**/*.html',
        '!deploy/**/*',
        '!node_modules/**/*'
    ]).pipe(gulp.dest(deployDir));
});

gulp.task('images', function() {
    return gulp.src('./images/*')
        .pipe(gulp.dest(path.join(deployDir, 'images')));
});


// watch for changes
gulp.task('watch', function() {
    gulp.watch(['./**/*.html', '!./deploy/**/*'], ['html']);
    gulp.watch('./images/**/*', ['images']);
    gulp.watch(['./scss/**/*.scss'], ['devCss']);
    gulp.watch(['./js/**/*.js'], ['devJs']);
});

gulp.task('dev', ['html', 'images', 'devCss', 'devJs']);

// start default task
gulp.task('default', ['html', 'images', 'css', 'js']);
