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

var assetDir = '../files/assets/';

var jsFiles = [
    'login.js',
    'index.js',
    'signup.js',
    'about.js',
    'first_run.js',
    'profile.js',
    'profile_edit.js',
    'admin.js',
    'editor.js',
];

var staticJSFiles = [{
    dir: './node_modules/vue/dist/',
    prod: 'vue.min.js',
    dev: 'vue.js',
    out: 'vue.js',
}, {
    dir: './node_modules/quill/dist/',
    prod: 'quill.min.js',
    dev: 'quill.js',
    out: 'quill.js',
}, ];

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
                file: path.join(assetDir, file),
                format: 'iife',
                sourcemap: sourcemaps,
            });
        });
}

function staticJS(dev) {
    let promises = [];
    for (let file of staticJSFiles) {
        promises.push(gulp.src(path.join(file.dir, dev ? file.dev : file.prod))
            .pipe(rename(file.out))
            .pipe(gulp.dest(path.join(assetDir, 'js')))
        );
    }

    return Promise.all(promises);
}


gulp.task('js', function() {
    return [
        rollupFiles(),
        staticJS(),
    ];
});

gulp.task('devJs', function() {
    return [
        rollupFiles(true),
        staticJS(true),
    ];
});

gulp.task('css', function() {
    return [gulp.src('./scss/**/*.scss')
        .pipe(sass({
            outputStyle: 'compressed',
            includePaths: 'node_modules'
        }).on('error', sass.logError))
        .pipe(postcss([
            autoprefixer({
                cascade: false
            })
        ]))
        .pipe(gulp.dest(path.join(assetDir, 'css'))),
        gulp.src('./node_modules/quill/dist/quill.snow.css')
        .pipe(gulp.dest(path.join(assetDir, 'css'))),
    ];
});

gulp.task('devCss', function() {
    return [gulp.src('./scss/**/*.scss')
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
        .pipe(gulp.dest(path.join(assetDir, 'css'))),
		gulp.src('./node_modules/quill/dist/quill.snow.css')
        .pipe(gulp.dest(path.join(assetDir, 'css'))),
	];
});


// static files
gulp.task('html', function() {
    return gulp.src([
        './**/*.html',
        '!node_modules/**/*'
    ]).pipe(gulp.dest(assetDir));
});

gulp.task('images', function() {
    return gulp.src('./images/*')
        .pipe(gulp.dest(path.join(assetDir, 'images')));
});

gulp.task('icons', function() {
    return gulp.src('./icons/*')
        .pipe(gulp.dest(path.join(assetDir, 'icons')));
});

// watch for changes
gulp.task('watch', function() {
    gulp.watch(['./**/*.html'], ['html']);
    gulp.watch('./images/**/*', ['images']);
    gulp.watch('./icons/**/*', ['icons']);
    gulp.watch(['./scss/**/*.scss'], ['devCss']);
    gulp.watch(['./js/**/*.js'], ['devJs']);
});

gulp.task('dev', ['html', 'images', 'icons', 'devCss', 'devJs']);

// start default task
gulp.task('default', ['html', 'images', 'icons', 'css', 'js']);
