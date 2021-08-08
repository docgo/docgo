import {nodeResolve} from '@rollup/plugin-node-resolve';
import scss from 'rollup-plugin-scss';
import { terser } from 'rollup-plugin-terser'

export default {
  input: 'index.js',
  output: {
    file: 'bundle.js',
    format: 'iife',
    name: "app"
  },
  plugins: [
    scss({
      failOnError: true,
      output: 'bundle.css'
    }),
    nodeResolve(),
    terser({ compress: { passes: 10 }, ecma: 2015, format: {ecma: 2015, comments: false, indent_level: 0} }),
  ]
}
