import React from 'react'
import { Tag } from 'antd'

interface MultipleTagsState {
  showall: boolean
}

interface MultipleTagsProps {
  arr: Array<string>
  keyPrefix?: string
  max?: number
}

class MultipleTags extends React.Component<MultipleTagsProps, MultipleTagsState> {
  state = {
    showall: false,
  }
  constructor(props: any) {
    super(props)
    this.handleShowMore = this.handleShowMore.bind(this)
    this.handleShowLess = this.handleShowLess.bind(this)
  }
  handleShowMore() {
    this.setState({ showall: true })
  }
  handleShowLess() {
    this.setState({ showall: false })
  }
  render() {
    let { arr, max, keyPrefix } = this.props
    let { showall } = this.state
    let tags: Array<JSX.Element> = []
    let maxTag = max ? max : 3
    let moreTagAdded = false
    if (arr) {
      arr.forEach(
        (value: string) => {
          if (showall || tags.length < maxTag) {
            tags.push(<Tag key={`${keyPrefix}${value}`}>{value}</Tag>)
          } else {
            if (!moreTagAdded) {
              moreTagAdded = true
              tags.push(<Tag color="blue" onClick={this.handleShowMore} key={`${keyPrefix}more`}>{"more"}</Tag>)
            }
          }
        }
      )
    }
    
    if (showall) {
      tags.push(<Tag color="blue" onClick={this.handleShowLess} key={`${keyPrefix}more`}>{"<"}</Tag>)
    }
    return (
    <div>
      {tags}
    </div>
    )
  }
}
export default MultipleTags
