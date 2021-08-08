import { nodeResolve } from '@rollup/plugin-node-resolve';
import scss from 'rollup-plugin-scss';

export default {
  input: 'index.js',
  output: {
    file: 'bundle.js',
    format: 'iife',
  },
  plugins: [
    scss({
        failOnError: true,
        output: 'bundle.css' 
    }),
    nodeResolve(),
  ]
}
