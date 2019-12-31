// Copyright 2019 Samaritan Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from 'react'
import {Tag} from 'antd'

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
    };

    constructor(props: any) {
        super(props);
        this.handleShowMore = this.handleShowMore.bind(this);
        this.handleShowLess = this.handleShowLess.bind(this)
    }

    handleShowMore() {
        this.setState({showall: true})
    }

    handleShowLess() {
        this.setState({showall: false})
    }

    render() {
        let {arr, max, keyPrefix} = this.props;
        let {showall} = this.state;
        let tags: Array<JSX.Element> = [];
        let maxTag = max ? max : 3;
        let moreTagAdded = false;
        if (arr) {
            arr.forEach(
                (value: string) => {
                    if (showall || tags.length < maxTag) {
                        tags.push(<Tag key={`${keyPrefix}${value}`}>{value}</Tag>)
                    } else {
                        if (!moreTagAdded) {
                            moreTagAdded = true;
                            tags.push(<Tag color="blue" onClick={this.handleShowMore}
                                           key={`${keyPrefix}more`}>{"more"}</Tag>)
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
