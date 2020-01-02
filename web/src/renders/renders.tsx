import React from 'react'
import { } from 'antd'
import MultipleTags from '../components/MultipleTags'

export function RenderStringArrayAsTags(input: string[]): JSX.Element {
  return (<MultipleTags arr={input}/>)
}